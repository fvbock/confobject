# ConfObject

Reloadable, validated configuration from many sources

[![Build Status](https://travis-ci.org/fvbock/confobject.png)](https://travis-ci.org/fvbock/confobject) [![GoDoc](https://godoc.org/github.com/fvbock/confobject?status.svg)](https://godoc.org/github.com/fvbock/confobject)


## Motivation

I had to deal with configuration values that came from a lot of different places and circumstances:

* a config file or env vars
* some json or yaml file - possibly stored in a db
* some of these i had control over - others could be modified by a user in an app that i do not necessarily must have access to to employ validation or control the way things are stored

I also needed:

* default values, "NOT NULL" like constraints, and validation that involves dependencies between different config values
* the ability to reload the configs from all its sources in its original order in an easy way

## Features

## Usage

	Cfg := struct {
		Config

		BoolSetting   bool    `default:"true"`
		StringSetting string  `default:"foo" json:"string_setting"`
		IntSetting    int     `required:"false" default:"23" should:"BeGreaterThanOrEqualTo_FloatSetting"`
		FloatSetting  float64 `default:"1.681"`

		SingleValueConfig struct {
			StringSetting string  `default:"bar" json:"single_value_config__string_setting"`
			IntSetting    int     `required:"false" default:"42" should:"BeGreaterThanOrEqualTo_.FloatSetting"`
			FloatSetting  float64 `should:"BeLessThanOrEqualTo_IntSetting"`
		}

		SliceConfig struct {
			StringSliceSetting []string  `default:"foo,bar"`
			IntSliceSetting    []int     `default:"23,42"`
			FloatSliceSetting  []float64 `default:"1.394,1.112"`
		}

		NestedValueConfig struct {
			StringSetting          string `default:"NestedOuterFoo"`
			InnerNestedValueConfig struct {
				StringSetting string `default:"NestedInnerFoo"`
			}
		}

		StringSet *set.StringSet `alias:someNameICantControl`
		IntSet    *set.IntSet
	}{}

	err = cobj.InitConfig(&Cfg, []cobj.InitFunc{
		cobj.InitFunc{
			F:           loadConfFile,
			ExitOnError: true,
		},
		cobj.InitFunc{
			F:           setupThisOtherThing,
			ExitOnError: false,
		},
}...)

## Limitations

## TODOs

* finish writing this file
* godoc
* clean up and extend the tests

## Questions?

Ping me [on twitter](https://twitter.com/fvbock)
