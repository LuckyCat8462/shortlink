// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	// 此处rest.RestConf才是结构体嵌入，代表config的实例可以直接访问所有restconf的字段和方法
	rest.RestConf
	// 此处hortUrlDB是一个匿名结构体（没有单独定义类型名）。
	// 这种写法是组合而非嵌入，Config 包含一个 ShortUrlDB 字段，
	// 但不能直接访问该结构体内的 DSN 字段（必须通过 config.ShortUrlDB.DSN）。
	ShortUrlDB struct {
		DSN string
	}

	SequenceDB struct {
		DSN string
	}
	ShortURLBlacklist []string
	ShortDomain       string
}
