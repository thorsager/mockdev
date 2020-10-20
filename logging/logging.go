package logging

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Trace(args ...interface{})
	Tracef(format string, args ...interface{})
	Traceln(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Warningln(args ...interface{})
	Warnln(args ...interface{})
}

//
////DefaultLogger is a logger warps logrus
//type DefaultLogger struct {
//	*logrus.Logger
//}
//
////NewDefaultLogger makes a new DefaultLogger
//func NewDefaultLogger() Logger {
//	var log = logrus.New()
//	log.Out = os.Stdout
//	log.Level = logrus.InfoLevel
//	return WrapLogrus(log)
//}
//
////WrapLogrus wraps a new DefaultLogger
//func WrapLogrus(p *logrus.Logger) Logger {
//	return &DefaultLogger{p}
//}
