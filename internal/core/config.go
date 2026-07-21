package core

import "github.com/sorrtory/freebooru/internal/config"

func (c *Core) ReadConfig() {
	cp, err := config.NewConfigParser(c.log)
	if err != nil {
		c.Fatal(err, "Failed to create config parser")
	}

	appConfigParser, err := config.NewAppConfigParser(c.log, c.opt.configFile)
	if err != nil {
		c.Fatal(err, "Failed to create app config parser")
	}

	appConfig, err := cp.Parse(appConfigParser)
	if err != nil {
		c.Fatal(err, "Failed to parse app config")
	}

}
