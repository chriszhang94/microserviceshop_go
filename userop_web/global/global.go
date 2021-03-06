package global

import (
	"mxshop/userop_web/proto"
	"mxshop/userop_web/config"
)

var (
	ServerConfig *config.ServerConfig = &config.ServerConfig{}
	NacosConfig *config.InitConfig = &config.InitConfig{}
	GoodsSrvClient proto.GoodsClient
	MessageClient proto.MessageClient
	AddressClient proto.AddressClient
	UserFavClient proto.UserFavClient
)
