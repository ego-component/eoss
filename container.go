package awos

import (
	"fmt"

	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
)

type Container struct {
	config *config
	name   string
	logger *elog.Component
}

func DefaultContainer() *Container {
	return &Container{
		config: DefaultConfig(),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

func Load(key string) *Container {
	c := DefaultContainer()
	if err := econf.UnmarshalKey(key, &c.config.bucketConfig); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	if err := econf.UnmarshalKey(key, &c.config); err != nil {
		c.logger.Panic("parse config error", elog.FieldErr(err), elog.FieldKey(key))
		return c
	}
	c.logger = c.logger.With(elog.FieldComponentName(key))
	c.name = key
	return c
}

func (c *Container) Build(options ...BuildOption) Component {
	for _, option := range options {
		option(c)
	}
	if c.config.bucketKey != "" {
		key := fmt.Sprintf("%s.buckets.%s", c.name, c.config.bucketKey)
		if err := econf.UnmarshalKey(key, &c.config.bucketConfig); err != nil {
			c.logger.Panic("parse bucket config error", elog.FieldErr(err), elog.FieldKey(key))
			return nil
		}
	}
	comp, err := newComponent(c.config)
	if err != nil {
		c.logger.Panic("new awos client fail", elog.FieldErr(err), elog.FieldValueAny(c.config))
	}
	return comp
}
