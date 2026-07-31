package config

var ExtConfig Extend

// Extend 扩展配置
//
//	extend:
//	  business:
//	    enabled: true
//	    driver: mysql
//	    source: user:password@tcp(127.0.0.1:3306)/business?charset=utf8&parseTime=True&loc=Local&timeout=1000ms
//
// 使用方法：config.ExtConfig.Business
type Extend struct {
	AMap     AMap       `yaml:"amap"`
	Business BusinessDB `yaml:"business"`
}

type AMap struct {
	Key string `yaml:"key"`
}

// BusinessDB 业务库配置（仅用于业务数据查询，不参与权限/租户分库）
type BusinessDB struct {
	Enabled         bool   `yaml:"enabled"`
	Driver          string `yaml:"driver"`
	Source          string `yaml:"source"`
	MaxIdleConns    int    `yaml:"maxIdleConns"`
	MaxOpenConns    int    `yaml:"maxOpenConns"`
	ConnMaxIdleTime int    `yaml:"connMaxIdleTime"`
	ConnMaxLifeTime int    `yaml:"connMaxLifeTime"`
}
