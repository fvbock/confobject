package confobject

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/fvbock/uds-go/set"
	cv "github.com/smartystreets/goconvey/convey"
	yaml "gopkg.in/yaml.v2"
)

var (
// Cfg *struct{}
)

func TestMain(m *testing.M) {
	fmt.Println("TestMain\n")

	TestMode = true

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

		StringSet *set.StringSet
		IntSet    *set.IntSet
	}{}

	err := InitConfig(&Cfg)
	if err != nil {
		fmt.Println(err)
	}

	// err = InitConfig(&Cfg)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println("*************")

	// fmt.Println(Cfg.Initialized)
	// fmt.Println(Cfg.ConfigKeys.Members())
	// fmt.Println(Cfg.ConfigTypes)
	fmt.Println(Cfg.StringSet.Members())

	// fmt.Println("*************")

	fmt.Println(Cfg)

	fmt.Println("*************")

	fmt.Println(Cfg.BoolSetting)
	fmt.Println(Cfg.StringSetting)
	fmt.Println(Cfg.IntSetting)
	fmt.Println(Cfg.FloatSetting)

	fmt.Println(Cfg.SingleValueConfig.StringSetting)
	fmt.Println(Cfg.SingleValueConfig.IntSetting)
	fmt.Println(Cfg.SingleValueConfig.FloatSetting)

	fmt.Println(Cfg.SliceConfig.StringSliceSetting)
	fmt.Println(Cfg.SliceConfig.IntSliceSetting)
	fmt.Println(Cfg.SliceConfig.FloatSliceSetting)

	fmt.Println(Cfg.NestedValueConfig.StringSetting)
	fmt.Println(Cfg.NestedValueConfig.InnerNestedValueConfig.StringSetting)

	fmt.Println("+++++++++++++ set with struct")

	// stuffs := struct {
	// 	StringSetting string
	// 	IntSetting    int
	// 	FloatSetting  float64
	// }{
	// 	"struct string",
	// 	23,
	// 	3.14,
	// }

	Cfg.BoolSetting = false
	fmt.Println(Cfg.BoolSetting)

	stuffs := struct {
		BoolSetting       bool
		StringSetting     string
		IntSetting        int
		FloatSetting      float64
		SingleValueConfig struct {
			StringSetting string
			IntSetting    int
			FloatSetting  float64
		}
	}{
		true,
		"struct string",
		23,
		3.14,
		struct {
			StringSetting string
			IntSetting    int
			FloatSetting  float64
		}{
			"SingleValueConfig struct string",
			1123,
			113.14,
		},
	}

	fmt.Println(stuffs)
	Cfg.Set(stuffs)

	fmt.Println(Cfg.BoolSetting)
	fmt.Println(Cfg.StringSetting)
	fmt.Println(Cfg.IntSetting)
	fmt.Println(Cfg.FloatSetting)

	fmt.Println(Cfg.SingleValueConfig.StringSetting)
	fmt.Println(Cfg.SingleValueConfig.IntSetting)
	fmt.Println(Cfg.SingleValueConfig.FloatSetting)

	fmt.Println("+++++++++++++ stuffs2")

	stuffs2 := map[string]string{
		"StringSetting": "map string string",
		"IntSetting":    "34",
		"FloatSetting":  "42.123",
	}

	// fmt.Println(stuffs2)
	Cfg.Set(stuffs2)

	fmt.Println("+++++++++++++")

	fmt.Println(Cfg.StringSetting)
	fmt.Println(Cfg.IntSetting)
	fmt.Println(Cfg.FloatSetting)

	fmt.Println(Cfg.SingleValueConfig.StringSetting)
	fmt.Println(Cfg.SingleValueConfig.IntSetting)
	fmt.Println(Cfg.SingleValueConfig.FloatSetting)

	fmt.Println("+++++++++++++ stuffs2a")

	stuffs2a := map[string]interface{}{
		"StringSetting": "map string interface",
		"IntSetting":    "45",
		"FloatSetting":  "53.123",
		"SingleValueConfig": map[string]interface{}{
			"StringSetting": "SingleValueConfig map string interface",
			"IntSetting":    "67",
			"FloatSetting":  "64.123",
		},
	}

	// fmt.Println(stuffs2a)
	Cfg.Set(stuffs2a)

	fmt.Println(Cfg.StringSetting)
	fmt.Println(Cfg.IntSetting)
	fmt.Println(Cfg.FloatSetting)

	fmt.Println(Cfg.SingleValueConfig.StringSetting)
	fmt.Println(Cfg.SingleValueConfig.IntSetting)
	fmt.Println(Cfg.SingleValueConfig.FloatSetting)

	fmt.Println("+++++++++++++ stuffs3")

	stuffs3 := map[string]map[string]string{
		"SingleValueConfig": map[string]string{
			"StringSetting": "map string map string",
			"IntSetting":    "89",
			"FloatSetting":  "75.14",
		},
	}

	fmt.Println(stuffs3)
	Cfg.Set(stuffs3)

	fmt.Println(Cfg.SingleValueConfig.StringSetting)
	fmt.Println(Cfg.SingleValueConfig.IntSetting)
	fmt.Println(Cfg.SingleValueConfig.FloatSetting)

	fmt.Println("+++++++++++++ stuffs4")

	stuffs4 := map[string][]string{
		"SliceConfig": []string{
			"StringSetting", "map slice string",
			"IntSetting", "map slice 42",
			"FloatSetting", "map slicestring 3.14",
		},
	}

	fmt.Println(stuffs4)
	Cfg.Set(stuffs4)

	fmt.Println(Cfg.SliceConfig.StringSliceSetting)
	fmt.Println(Cfg.SliceConfig.IntSliceSetting)
	fmt.Println(Cfg.SliceConfig.FloatSliceSetting)

	fmt.Println("+++++++++++++ [][]interface")

	var stuffs5 = [][]interface{}{
		[]interface{}{
			"SliceConfig.StringSliceSetting", "foo2", "foo3",
		},
		[]interface{}{
			"SliceConfig.IntSliceSetting", 232323, 424242,
		},
		[]interface{}{
			"SliceConfig.FloatSliceSetting", 2.32323, 4.24242,
		},
	}
	Cfg.Set(stuffs5)

	fmt.Println("5>", Cfg.SliceConfig.StringSliceSetting)
	fmt.Println("5>", Cfg.SliceConfig.IntSliceSetting)
	fmt.Println("5>", Cfg.SliceConfig.FloatSliceSetting)

	fmt.Println("+++++++++++++ [][]string")

	var stuffs5s = [][]string{
		[]string{
			"SliceConfig.StringSliceSetting", "foobaz",
		},
		[]string{
			"SliceConfig.IntSliceSetting", "123456", "7890",
		},
		[]string{
			"SliceConfig.FloatSliceSetting", "123.456", "78.90",
		},
	}
	Cfg.Set(stuffs5s)

	fmt.Println("5s>", Cfg.SliceConfig.StringSliceSetting)
	fmt.Println("5s>", Cfg.SliceConfig.IntSliceSetting)
	fmt.Println("5s>", Cfg.SliceConfig.FloatSliceSetting)

	ret := m.Run()
	log.Println(ret)
	os.Exit(ret)
}

func TestDefaultValues(t *testing.T) {
	cfg := struct {
		Config

		BoolSetting   bool    `default:"true"`
		StringSetting string  `default:"foo"`
		IntSetting    int     `default:"23"`
		FloatSetting  float64 `default:"1.681"`

		SingleValueConfig struct {
			StringSetting string  `default:"bar"`
			IntSetting    int     `default:"42"`
			FloatSetting  float64 `default:"23.12"`
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

		StringSet *set.StringSet
		IntSet    *set.IntSet
	}{}

	initFunc := InitFunc{
		F: func() (err error) {
			cfg.StringSet = set.NewStringSet([]string{"foo", "bar"}...)
			cfg.IntSet = set.NewIntSet([]int{1, 2}...)
			return
		},
		ExitOnError: false,
	}

	err := InitConfig(&cfg, initFunc)
	cv.Convey(`Initializing the config should pass.`, t, func() {
		cv.So(err, cv.ShouldBeNil)

		cv.So(cfg.BoolSetting, cv.ShouldBeTrue)
		cv.So(cfg.StringSetting, cv.ShouldEqual, "foo")
		cv.So(cfg.IntSetting, cv.ShouldEqual, 23)
		cv.So(cfg.FloatSetting, cv.ShouldEqual, 1.681)

		cv.So(cfg.SingleValueConfig.StringSetting, cv.ShouldEqual, "bar")
		cv.So(cfg.SingleValueConfig.IntSetting, cv.ShouldEqual, 42)
		cv.So(cfg.SingleValueConfig.FloatSetting, cv.ShouldEqual, 23.12)

		cv.So(cfg.SliceConfig.StringSliceSetting, cv.ShouldContain, "foo")
		cv.So(cfg.SliceConfig.StringSliceSetting, cv.ShouldContain, "bar")
		cv.So(cfg.SliceConfig.IntSliceSetting, cv.ShouldContain, 23)
		cv.So(cfg.SliceConfig.IntSliceSetting, cv.ShouldContain, 42)
		cv.So(cfg.SliceConfig.FloatSliceSetting, cv.ShouldContain, 1.394)
		cv.So(cfg.SliceConfig.FloatSliceSetting, cv.ShouldContain, 1.112)

		cv.So(cfg.NestedValueConfig.StringSetting, cv.ShouldEqual, "NestedOuterFoo")
		cv.So(cfg.NestedValueConfig.InnerNestedValueConfig.StringSetting, cv.ShouldEqual, "NestedInnerFoo")

		cv.So(cfg.StringSet.HasMembers("foo", "bar"), cv.ShouldBeTrue)
		cv.So(cfg.IntSet.HasMembers(1, 2), cv.ShouldBeTrue)
	})

}

func TestEmptyAndZeroDefaultAssertions(t *testing.T) {
	var err error

	cfg1 := struct {
		Config

		StringDefaultNotEmpty string `required:"true" should:"NotBeBlank"`
	}{}
	err = InitConfig(&cfg1)
	log.Println(cfg1.Initialized, cfg1.StringDefaultNotEmpty, len(cfg1.StringDefaultNotEmpty))
	cv.Convey(`Initializing a config with a required should:"NotBeEmpty" string value should fail initialization.`, t, func() {
		cv.So(err, cv.ShouldNotBeNil)
	})
	cfg1.StringDefaultNotEmpty = "foo"
	err = InitConfig(&cfg1)
	cv.Convey(`setting the value should make initialization pass.`, t, func() {
		cv.So(err, cv.ShouldBeNil)
	})

	cfg2 := struct {
		Config

		IntRequiredGTZero int `required:"true" should:"BeGreaterThan:0"`
	}{}
	err = InitConfig(&cfg2)
	cv.Convey(`Initializing a config with a required should:"BeGreaterThan:0" int value should fail initialization.`, t, func() {
		cv.So(err, cv.ShouldNotBeNil)
	})
	cfg2.IntRequiredGTZero = 42
	err = InitConfig(&cfg2)
	cv.Convey(`setting the value should make initialization pass.`, t, func() {
		cv.So(err, cv.ShouldBeNil)
	})

	cfg3 := struct {
		Config

		FloatRequiredGTEZero float64 `required:"true" should:"BeGreaterThan:0"`
	}{}
	err = InitConfig(&cfg3)
	cv.Convey(`Initializing a config with a required should:"BeGreaterThan:0" float value should fail initialization.`, t, func() {
		cv.So(err, cv.ShouldNotBeNil)
	})
	cfg3.FloatRequiredGTEZero = 42.23
	err = InitConfig(&cfg3)
	cv.Convey(`setting the value should make initialization pass.`, t, func() {
		cv.So(err, cv.ShouldBeNil)
	})

	cfg4 := struct {
		Config

		RangeLow  float64 `required:"true" should:"BeGreaterThan:0"`
		RangeHigh float64 `required:"true" should:"BeGreaterThan_RangeLow BeLessThan:10"`
	}{}
	cfg4.RangeLow = 1
	err = InitConfig(&cfg4)
	cv.Convey(`setting RangeLow to 1 and not setting RangeHigh should fail initialization.`, t, func() {
		cv.So(err, cv.ShouldNotBeNil)
	})

	cfg4.RangeHigh = 0
	err = InitConfig(&cfg4)
	cv.Convey(`setting RangeHigh then to 0 should still fail initialization.`, t, func() {
		cv.So(err, cv.ShouldNotBeNil)
	})

	cfg4.RangeHigh = 11
	err = InitConfig(&cfg4)
	cv.Convey(`setting RangeHigh then to 11 should still fail initialization.`, t, func() {
		cv.So(err, cv.ShouldNotBeNil)
	})

	cfg4.RangeHigh = 8
	err = InitConfig(&cfg4)
	cv.Convey(`setting RangeHigh then to 8 should then pass initialization.`, t, func() {
		cv.So(err, cv.ShouldBeNil)
	})

	// cfg := struct {
	// 	Config

	// 	StringDefaultNotEmpty string  `required:"true" should:"NotBeBlank"`
	// 	IntRequiredGTZero     int     `required:"true" should:"BeGreaterThan:0"`
	// 	FloatRequiredGTEZero  float64 `required:"true" should:"BeGreaterThan:0"`
	// 	RangeLow              float64 `required:"true" should:"BeGreaterThan:0"`
	// }{}

	// err := InitConfig(&cfg)
	// if err != nil {
	// 	t.Log(err)
	// 	t.Fail()
	// } else {
	// 	t.Log(cfg.RangeLow)
	// 	t.Log(cfg)
	// }

}

func TestSetFromEnv(t *testing.T) {
	var err error
	os.Setenv(fmt.Sprintf("%s%s", ENV_PREFIX, "StringDefaultNotEmpty"), "FooFromEnv")
	cfg := struct {
		Config

		StringDefaultNotEmpty string `required:"true" should:"NotBeBlank"`
	}{}
	err = InitConfig(&cfg)
	cv.Convey(`Initializing a config with a required should:"NotBeEmpty" string value should pass initialization when the prefixed key is set on the OS.ENV.`, t, func() {
		cv.So(err, cv.ShouldBeNil)
		cv.So(cfg.StringDefaultNotEmpty, cv.ShouldEqual, "FooFromEnv")
	})
}

func TestMembershipAssertions(t *testing.T) {
	var err error

	cfg1 := struct {
		Config

		StringSetting string `default:"foo" should:"BeIn:foo,bar"`
		// IntSetting    int     `default:"32" should:"BeIn:23,42"`
		// FloatSetting  float64 `default:"1.861" should:"BeIn:1.681,3.141"`
	}{}

	err = InitConfig(&cfg1)
	cv.Convey(`Initializing the config should pass.`, t, func() {
		cv.So(err, cv.ShouldBeNil)

		cfg1.StringSetting = "ofo"
		err = cfg1.Validate()
		cv.So(err, cv.ShouldNotBeNil)

		cfg1.StringSetting = "bar"
		err = cfg1.Validate()
		cv.So(err, cv.ShouldBeNil)
		cv.So(cfg1.StringSetting, cv.ShouldEqual, "bar")
	})

	cfg2 := struct {
		Config

		IntSetting int `default:"23" should:"BeIn:23,42"`
		// FloatSetting  float64 `default:"1.861" should:"BeIn:1.681,3.141"`
	}{}

	err = InitConfig(&cfg2)
	cv.Convey(`Initializing the config should pass.`, t, func() {
		cv.So(err, cv.ShouldBeNil)

		cfg2.IntSetting = 32
		err = cfg2.Validate()
		cv.So(err, cv.ShouldNotBeNil)

		cfg2.IntSetting = 42
		err = cfg2.Validate()
		cv.So(err, cv.ShouldBeNil)
		cv.So(cfg2.IntSetting, cv.ShouldEqual, 42)
	})

	cfg3 := struct {
		Config

		FloatSetting float64 `default:"1.681" should:"BeIn:1.681,3.141"`
	}{}

	err = InitConfig(&cfg3)
	cv.Convey(`Initializing the config should pass.`, t, func() {
		cv.So(err, cv.ShouldBeNil)

		cfg3.FloatSetting = 1.861
		err = cfg3.Validate()
		cv.So(err, cv.ShouldNotBeNil)

		cfg3.FloatSetting = 3.141
		err = cfg3.Validate()
		cv.So(err, cv.ShouldBeNil)
		cv.So(cfg3.FloatSetting, cv.ShouldEqual, 3.141)
	})
}

func TestAliases(t *testing.T) {
}

func TestInitSets(t *testing.T) {
	var err error
	cfg1 := struct {
		Config

		StringSet *set.StringSet `default:"foo,bar"`
		// StringSet *set.StringSet `default:"foo,bar" should:"BeIn:foo,bar,baz"`
	}{}

	err = InitConfig(&cfg1)
	cv.Convey(`Initializing the config should pass.`, t, func() {
		cv.So(err, cv.ShouldBeNil)
		cv.So(cfg1.StringSet.HasMembers("foo", "bar"), cv.ShouldBeTrue)
		t.Log(cfg1.StringSet.Members())
	})
}

func TestLoadYaml(t *testing.T) {
	var err error
	cfg := struct {
		Config

		BoolSetting   bool    `required:"true"`
		StringSetting string  `required:"true"`
		IntSetting    int     `required:"true"`
		FloatSetting  float64 `required:"true"`

		SliceConfig struct {
			StringSliceSetting []string  `required:"true"`
			IntSliceSetting    []int     `required:"true"`
			FloatSliceSetting  []float64 `required:"true"`
		}
	}{}

	// initFunc := func() (err error) {
	// 	var yamlData []byte
	// 	yamlData, err = ioutil.ReadFile("test/test.yaml")
	// 	if err != nil {
	// 		return
	// 	}
	// 	v := [][]string{}
	// 	err = yaml.Unmarshal([]byte(yamlData), &v)
	// 	if err != nil {
	// 		return
	// 	}
	// 	log.Println(v)
	// 	log.Println(cfg)

	// 	err = cfg.Set(v)
	// 	return
	// }

	cv.Convey(`Initializing the config should pass.`, t, func() {
		err = InitConfig(&cfg)
		// cv.So(err, cv.ShouldBeNil)

		var yamlData []byte
		yamlData, err = ioutil.ReadFile("test/test.yaml")
		cv.So(err, cv.ShouldBeNil)

		v := [][]string{}
		err = yaml.Unmarshal([]byte(yamlData), &v)
		cv.So(err, cv.ShouldBeNil)

		t.Log("---", v)
		// log.Println(cfg)

		err = cfg.Set(v)
		t.Log(cfg.SliceConfig.StringSliceSetting)
		t.Log(cfg.SliceConfig.IntSliceSetting)
		// cv.So(err, cv.ShouldBeNil)
	})

}
