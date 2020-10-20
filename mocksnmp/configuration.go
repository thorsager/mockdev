package mocksnmp

type Configuration struct {
	Name           string   `yaml:"name"`
	BindAddr       string   `yaml:"bind-addr"`
	SnapshotFiles  []string `yaml:"snapshot-files"`
	OIDs           []string `yaml:"oids"`
	ReadCommunity  string   `yaml:"community-ro"`
	WriteCommunity string   `yaml:"community-rw"`
}
