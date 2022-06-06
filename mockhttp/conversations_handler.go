package mockhttp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/thorsager/mockdev/logging"
	"github.com/thorsager/mockdev/rawhttp"
	"github.com/thorsager/mockdev/scripts"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

type ConversationsHandler struct {
	sync.Mutex
	Conversations      []Conversation
	Log                logging.Logger
	SessionLogLocation string
	SessionLogReceived bool
	sessionCounter     int
	BindAddress        string
}

func (h *ConversationsHandler) sessionContext() context.Context {
	h.Lock()
	defer h.Unlock()
	h.sessionCounter = h.sessionCounter + 1
	ctx := contextWithWithSessionId(h.sessionCounter)
	setContextLogger(ctx, h.Log)
	return ctx
}

func (h *ConversationsHandler) sessionFilename(sesId int) string {
	return path.Join(h.SessionLogLocation, fmt.Sprintf("sess_%.4d.log", sesId))
}

func (h *ConversationsHandler) initSessionLog(ctx context.Context) error {
	if h.SessionLogReceived {
		sesId, err := getSessionId(ctx)
		if err != nil {
			return err
		}
		filename := h.sessionFilename(sesId)
		h.Log.Debugf("logging session to: %s", filename)
		err = os.MkdirAll(h.SessionLogLocation, 0770)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = f.WriteString(fmt.Sprintf("# session %d, %s \n", sesId, time.Now()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *ConversationsHandler) logRequestBody(ctx context.Context, reader io.Reader) error {
	if !h.SessionLogReceived {
		return nil
	}
	sesId, err := getSessionId(ctx)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(h.sessionFilename(sesId), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(f, reader)
	if err != nil {
		return err
	}
	return nil
}

func (h *ConversationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.sessionContext()
	err := h.initSessionLog(ctx)
	if err != nil {
		h.Log.Errorf("sessionLogInit: %v", err)
		http.Error(w, "while initializing session logging: "+err.Error(), 418)
		return
	}

	// get body and re-install
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.Log.Errorf("readBody: %v", err)
		http.Error(w, "while reading request body: "+err.Error(), 418)
		return
	}
	_ = r.Body.Close() //  must close
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	h.Log.Tracef("Request %s %s", r.Method, r.URL)
	h.Log.Tracef("%s", bodyBytes)

	var theOne Conversation
	candidates, breaker := h.filterConversations(ctx, r)
	if breaker != nil {
		theOne = *breaker
		h.Log.Debugf("Breaking match on: %s", theOne.Name)
	} else {
		if len(candidates) < 1 {
			http.Error(w, "I'm not a teapot", 418)
			h.Log.Warnf("No matches after conversation-filter: %s \n%s", r.URL.Path, string(bodyBytes))
			return
		}
		score, _ := getConversationScores(ctx)
		h.Log.Debugf("scoreKey: %#v", score)
		theOne, err = score.tieBreak(candidates)
		if err != nil {
			http.Error(w, "I'm not a teapot", 418)
			h.Log.Warnf("No matches after conversation-scoring: %s \n%s", r.URL.Path, string(bodyBytes))
			return
		}
	}

	if err := handleDelay(theOne.Response.Delay); err != nil {
		h.Log.Errorf("While handling response-delay: %v", err)
	}

	_ = h.logRequestBody(ctx, bytes.NewBuffer(bodyBytes))
	_ = h.serveResponse(w, r, theOne)
}

func handleDelay(delay ResponseDelay) error {
	if delay.Max == 0 && delay.Min == 0 || os.Getenv("IGNORE_DELAY") != "" {
		return nil // no delay
	}
	interval := 0
	delta := delay.Max - delay.Min
	if delta < 0 {
		return fmt.Errorf("invalid sleep interval min=%d, max=%d", delay.Min, delay.Max)
	}
	if delta == 0 {
		interval = delay.Min
	} else {
		interval = rand.Intn(delta)
	}
	if interval > 0 {
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
	return nil
}

func (h *ConversationsHandler) filterConversations(ctx context.Context, r *http.Request) (candidates []Conversation, breaker *Conversation) {
	for _, conversation := range h.Conversations {
		h.Log.Debugf("Matching [%d] '%s'", conversation.Order, conversation.Name)
		methodMatch := matchMethod(ctx, r, conversation)
		urlMatch := matchURL(ctx, r, conversation)
		headersMatch := matchHeaders(ctx, r, conversation)
		bodyMatch := matchBody(ctx, r, conversation)

		allMatch := methodMatch && urlMatch && headersMatch && bodyMatch

		if conversation.BreakOnMatch() && allMatch {
			h.Log.Debugf("Breaking on 'match' '%s'", conversation.Name)
			return nil, &conversation
		} else if conversation.BreakOnNoMatch() && !allMatch {
			h.Log.Debugf("Breaking on 'no-match' '%s'", conversation.Name)
			return nil, &conversation
		}
		if allMatch {
			h.Log.Debugf("Matching all '%s'", conversation.Name)
			candidates = append(candidates, conversation)
		} else {
			h.Log.Tracef("Disregarding '%s' methodMatch=%t, urlMatch=%t, headerMatch=%t, bodyMatch=%t", conversation.Name, methodMatch, urlMatch, headersMatch, bodyMatch)
		}
	}
	return candidates, nil
}

func (h *ConversationsHandler) createBaseTemplateData() map[string]interface{} {
	td := make(templateData)
	td[cfg] = createConfigData(h.BindAddress)
	td[env] = createEnvData()

	now := time.Now()
	loc, _ := time.LoadLocation("GMT")
	td[currentTime] = now
	td[currentTimeGMT] = now.In(loc)
	return td
}

func (h *ConversationsHandler) serveResponse(w http.ResponseWriter, r *http.Request, conversation Conversation) error {

	templateVars := h.createBaseTemplateData()

	if conversation.Request.UrlMatcher.Path != "" {
		m := regexp.MustCompile(conversation.Request.UrlMatcher.Path)
		matches := m.FindStringSubmatch(r.URL.Path)
		for i, j := range matches {
			h.Log.Debugf("p%d=%s", i, j)
			templateVars[fmt.Sprintf("p%d", i)] = j
		}
	}

	if conversation.Request.UrlMatcher.Query != "" {
		m := regexp.MustCompile(conversation.Request.UrlMatcher.Query)
		matches := m.FindStringSubmatch(r.URL.RawQuery)
		for i, j := range matches {
			templateVars[fmt.Sprintf("q%d", i)] = j
		}
	}

	if conversation.Request.BodyMatcher != "" {
		// get body and re-install
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		_ = r.Body.Close() //  must close
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		m := regexp.MustCompile(conversation.Request.BodyMatcher)
		matches := m.FindSubmatch(bodyBytes)
		for i, j := range matches {
			h.Log.Tracef("match-group %d = '%s'", i, string(j))
			templateVars[fmt.Sprintf("b%d", i)] = string(j)
		}
	}
	// TODO: Figure out how match-groups could be implemented on Header Matchers.
	h.Log.Tracef("templateVars: %+v", templateVars)

	if len(conversation.Response.Script) > 0 {
		resp := h.executeScript(conversation.Response.Script, templateVars)
		h.Log.Tracef("Scripted response:\n--\n%s\n--\n", resp)
		_, err := rawhttp.NewWriter(w).Write(resp)
		if err != nil {
			return err
		}
	} else {
		bodyBuffer := &bytes.Buffer{}
		if conversation.Response.BodyFile != "" {
			f, err := os.Open(conversation.Response.BodyFile)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			_, err = io.Copy(bodyBuffer, f)
			if err != nil {
				return err
			}
		} else {
			bodyBuffer.Write([]byte(conversation.Response.Body))
		}

		// parse headers as templates
		executedBuffer := &bytes.Buffer{}
		for _, s := range conversation.Response.Headers {
			tmpl, err := template.New("body").Parse(s)
			if err != nil {
				return err
			}
			err = tmpl.Execute(executedBuffer, templateVars)
			addHeaderFromString(w, executedBuffer.String())
			executedBuffer.Reset()
		}

		// parse body as template
		tmpl, err := template.New("body").Parse(bodyBuffer.String())
		if err != nil {
			return err
		}

		err = tmpl.Execute(executedBuffer, templateVars)
		if err != nil {
			return err
		}

		w.Header().Add("X-Powered-By", "mockdev")
		w.Header().Add("size", fmt.Sprintf("%d", executedBuffer.Len()))
		w.WriteHeader(conversation.Response.StatusCode)
		_, _ = w.Write(executedBuffer.Bytes())
	}

	h.Log.Infof("Served response from conversation: '%d:%s' (%s)", conversation.Order, conversation.Name, r.URL)

	_ = h.executeScript(conversation.AfterScript, templateVars)
	return nil
}

func (h *ConversationsHandler) executeScript(script []string, templateVars map[string]interface{}) []byte {
	out := bytes.Buffer{}
	if len(script) > 0 {
		h.Log.Tracef("Script:\n%s\n", strings.Join(script, "\n"))
		// build env
		localEnv := make(map[string]string)
		for k, v := range templateVars {
			if k != env && k != cfg {
				if value, ok := v.(string); ok {
					localEnv[k] = value
				}
			}
		}
		h.Log.Tracef("env: %+v", localEnv)
		for _, line := range script {
			stdout, stderr, err := scripts.Execute(line, localEnv)
			h.Log.Tracef(">> %s\n", line)
			if err != nil {
				if len(stdout) > 0 {
					out.Write(stdout)
					h.Log.Tracef("<< %s", stdout)
				}
				if len(stderr) > 0 {
					h.Log.Warnf("%s", stderr)
				}
				h.Log.Errorf("%s", err)
				break
			} else {
				if len(stdout) > 0 {
					out.Write(stdout)
					h.Log.Tracef("<< %s", stdout)
				}
				if len(stderr) > 0 {
					h.Log.Warnf("%s", stderr)
				}
			}
		}
	}
	return out.Bytes()
}

func addHeaderFromString(w http.ResponseWriter, s string) {
	t := strings.SplitN(s, ":", 2)
	w.Header().Add(t[0], t[1])
}
