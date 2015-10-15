package confobject

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/fvbock/uds-go/set"
)

const (
	ENV_PREFIX = "CONFOBJ_"

	TAG_DEFAULT         = "default"
	TAG_REQUIRED        = "required"
	TAG_ALIAS           = "alias"
	TAG_ASSERTION       = "should"
	TAG_ASSERTION_SEP   = " "
	TAG_ASSERTION_VALUE = ":"
	TAG_ASSERTION_FIELD = "_"
)

type Config struct {
	MainConfig             reflect.Value
	ConfigValues           map[string]interface{}
	Initialized            bool
	ConfigKeys             *set.StringSet
	KeyAliases             *set.StringSet
	AliasKeyMap            map[string]string
	ConfigTypes            map[string]string
	ConfigTags             map[string]reflect.StructTag
	PanicOnAssignmentError bool
	Assertions             map[string]map[string][]Assertion
	initFuncs              []InitFunc
	LogSet                 bool
	// initFuncs              []func() (err error)
}

func (c Config) String() string {
	var repr = "Current Config:\n"
	for _, key := range c.ConfigKeys.Members() {
		fld, _ := c.FieldForKey(key)
		repr += fmt.Sprintf("%s: %v.\n", key, fld.Interface())
	}
	return repr
}

type InitFunc struct {
	F           func() (err error)
	ExitOnError bool
}

func InitConfig(c interface{}, initFuncs ...InitFunc) (err error) {
	cval := reflect.ValueOf(c)
	for cval.Kind() == reflect.Ptr {
		cval = cval.Elem()
	}
	cf := cval.FieldByName("Config").Interface().(Config)

	if cf.Initialized {
		err = fmt.Errorf("Config %v is already Initialized.", cf)
		return
	}

	names, types, tags := StructFields(c)
	namesSet := set.NewStringSet()
	assertions := make(map[string]map[string][]Assertion)
	for _, key := range names {
		if !namesSet.HasMember(key) {
			namesSet.Add(key)
			assertions[key] = make(map[string][]Assertion)
		} else {
			err = fmt.Errorf("Coniguration keys must be unique - got '%s' a second time.", key)
			fmt.Println(err)
			return
		}
	}

	// init sets
	for key, type_ := range types {
		switch type_ {
		case "*set.StringSet":
			cval.FieldByName(key).Set(reflect.ValueOf(set.NewStringSet()))
		case "*set.IntSet":
			cval.FieldByName(key).Set(reflect.ValueOf(set.NewIntSet()))
		case "*set.Int64Set":
			cval.FieldByName(key).Set(reflect.ValueOf(set.NewInt64Set()))
		}
	}

	initC := Config{
		MainConfig:   cval,
		ConfigValues: make(map[string]interface{}),
		Initialized:  false,
		ConfigKeys:   namesSet,
		KeyAliases:   set.NewStringSet(),
		AliasKeyMap:  make(map[string]string),
		ConfigTypes:  types,
		ConfigTags:   tags,
		Assertions:   assertions,
		LogSet:       cf.LogSet,
	}

	initC.initFuncs = append(
		[]InitFunc{
			InitFunc{
				F:           initC.setDefaults,
				ExitOnError: true,
			},
			InitFunc{
				F:           initC.readFromEnv,
				ExitOnError: true,
			},
		},
		append(
			initFuncs,
			InitFunc{
				F:           initC.Validate,
				ExitOnError: true,
			},
		)...,
	)

	// get aliases
	err = initC.extractAliases()
	if err == nil {
		// set Assertions
		err = initC.extractAssertions()
		if err == nil {
			log.Println("initC.extractAssertions() OK")
		}
	}

	// init functions might reference a global config - we now need to operate
	// on the actualy object..
	cval.FieldByName("Config").Set(reflect.ValueOf(initC))
	cfg := cval.FieldByName("Config").Interface().(Config)
	// run all init functions
	err = cfg.ReInit()
	if err == nil {
		log.Println("initC.ReInit() OK")
		cval.FieldByName("Initialized").SetBool(true)
	}

	return
}

func (c *Config) FieldForKey(key string) (field reflect.Value, err error) {
	field = c.MainConfig
	if !c.ConfigKeys.HasMember(key) {
		if _, notSet := c.AliasKeyMap[key]; !notSet {
			err = fmt.Errorf("Unknown config key or alias: %s", key)
			return
		}
		key = c.AliasKeyMap[key]
	}
	keyParts := strings.Split(key, ".")
	for _, k := range keyParts {
		field = field.FieldByName(k)
	}
	return
}

func (c *Config) ReInit() (err error) {
	for _, f := range c.initFuncs {
		err = f.F()
		fName := strings.Replace(
			runtime.FuncForPC(reflect.ValueOf(f.F).Pointer()).Name(),
			"github.com/fvbock/confobject.",
			"",
			1,
		)
		if err != nil {
			log.Printf("INIT: %s ERROR:%v\n", fName, err)
			if f.ExitOnError {
				os.Exit(1)
			}
			return
		} else {
			log.Printf("INIT: %s ok.\n", fName)
		}
	}
	return
}

func (c *Config) Validate() (err error) {
	// assertionsLoop:
	for key, assertionsMap := range c.Assertions {
		for targetKey, assertions := range assertionsMap {
			// log.Println("Validate() >>", key, targetKey, assertions)
			for _, asrtn := range assertions {
				var isMemberAssertion = false
				assertionName := runtime.FuncForPC(reflect.ValueOf(asrtn).Pointer()).Name()
				if len(assertionName) >= 6 &&
					(assertionName[len(assertionName)-4:] == "BeIn" ||
						assertionName[len(assertionName)-4:] == "NotBeIn") {
					isMemberAssertion = true
				}

				fld, err := c.FieldForKey(key)
				if err != nil {
					log.Println(err)
					continue
				}
				if c.ConfigTypes[key] == "*set.StringSet" {
					fld = reflect.ValueOf(fld.Interface().(*set.StringSet).Members())
				}

				if key == targetKey {
					// log.Println("src, target same")
					if ok, message := Assert(fld.Interface(), asrtn); !ok {
						return fmt.Errorf("Assertion failure on %s: %s", key, message)
					}
					// log.Println("ok")
					continue
				}

				var targetFld reflect.Value
				// if we dont know the key it is a literal value.
				// existance is checked in extractAssertions
				if c.ConfigKeys.HasMember(targetKey) {
					targetFld, err = c.FieldForKey(targetKey)
					if err != nil {
						log.Println(err)
						return err
					}
				} else {
					// literal
					switch c.ConfigTypes[key] {
					case "string":
						if !isMemberAssertion {
							targetFld = reflect.ValueOf(targetKey)
						} else {
							var type_ string
							strSlice, err := sliceFromStrings(targetKey, type_)
							if err != nil {
								return err
							}
							targetFld = reflect.ValueOf(strSlice.([]string))
						}

					case "int":
						if !isMemberAssertion {
							intval, err := intFromInterface(targetKey)
							if err != nil {
								log.Println(err)
								return err
							}
							targetFld = reflect.ValueOf(int(intval))
						} else {
							var type_ int
							intSlice, err := sliceFromStrings(targetKey, type_)
							if err != nil {
								return err
							}

							targetFld = reflect.ValueOf(intSlice.([]int))
						}

					case "float64":
						if !isMemberAssertion {
							fval, err := floatFromInterface(targetKey)
							if err != nil {
								log.Println(err)
								return err
							}
							targetFld = reflect.ValueOf(float64(fval))
						} else {
							var type_ float64
							floatSlice, err := sliceFromStrings(targetKey, type_)
							if err != nil {
								return err
							}

							targetFld = reflect.ValueOf(floatSlice.([]float64))
						}
					}
				}

				if ok, message := Assert(fld.Interface(), asrtn, targetFld.Interface()); !ok {
					return fmt.Errorf("Assertion failure: %s", message)
				}

				// log.Println("ok")
			}
		}
	}
	log.Println("Config Validation OK")
	return
}

func (c *Config) Set(configData interface{}, prependKeys ...string) (err error) {
	// if !c.Initialized {
	// 	err = fmt.Errorf("Cannot set any values on an uninitialized Object. Call InitConfig() first!")
	// }
	configValue := reflect.ValueOf(configData)
	switch configValue.Kind() {
	case reflect.Struct:
		_, types, _ := StructFields(configData)

		for key, _ := range types {
			field := configValue
			keyParts := strings.Split(key, ".")
			for _, k := range keyParts {
				field = field.FieldByName(k)
			}

			err = c.setValue(key, field.Interface())
			if err != nil {
				// panic(err.Error())
				log.Println(err)
			}
		}

	case reflect.Map:
		for _, keyVal := range configValue.MapKeys() {
			key := keyVal.Interface().(string)
			// fmt.Println(key, "kind", reflect.ValueOf(configValue.MapIndex(keyVal).Interface()).Kind())
			switch reflect.ValueOf(configValue.MapIndex(keyVal).Interface()).Kind() {
			case reflect.Map, reflect.Slice:
				prependKeys = append(prependKeys, key)
				c.Set(configValue.MapIndex(keyVal).Interface(), prependKeys...)
				continue
			case reflect.Struct:
				// TODO
			}
			if len(prependKeys) > 0 {
				key = fmt.Sprintf("%s.%s", strings.Join(prependKeys, "."), key)
			}
			// log.Println("---", key, configValue.MapIndex(keyVal).Interface())
			err = c.setValue(key, configValue.MapIndex(keyVal).Interface())
			if err != nil {
				// panic(err.Error())
				log.Println(err)
				// return
			}
		}

	case reflect.Slice:
		if len(prependKeys) == 0 {
			for i := 0; i < configValue.Len(); i++ {
				if i == 0 {
					keyCandidate, ok := configValue.Index(i).Interface().(string)
					// fmt.Println("ok?", ok, c.ConfigKeys.HasMember(keyCandidate), c.KeyAliases.HasMember(keyCandidate))
					if ok && (c.ConfigKeys.HasMember(keyCandidate) ||
						c.KeyAliases.HasMember(keyCandidate)) {
						err = c.setValue(keyCandidate, configValue.Slice(1, configValue.Len()).Interface())
						if err != nil {
							// panic(err.Error())
							log.Println(err)
						}
						return
					}
				}

				err = c.Set(configValue.Index(i).Interface())
				if err != nil {
					// panic(err.Error())
					log.Println(err)
				}
			}
		} else {
			key := fmt.Sprintf("%s", strings.Join(prependKeys, "."))
			if c.ConfigKeys.HasMember(key) {
				err = c.setValue(key, configValue.Interface())
			}
		}
	default:
		// panic(fmt.Sprintf("I got stuff i can't deal with: %v\n", configData))
		log.Printf("I got stuff i can't deal with: %v\n", configData)
	}

	return
}

func (c *Config) setValue(key string, value interface{}) (err error) {
	var field reflect.Value
	// fmt.Println("%", key, value)
	field, err = c.FieldForKey(key)
	if err != nil {
		return
	}

	ifaces, ok := value.([]interface{})
	if ok {
		fmt.Println("[]interface{}...")
		var val interface{}
		for _, iface := range ifaces {
			// fmt.Println("> got", reflect.TypeOf(iface))
			switch c.ConfigTypes[key] {
			case "bool":
				val = iface.(bool)
			case "[]bool":
				if reflect.TypeOf(val) == nil {
					val = []bool{}
				}
				val = append(val.([]bool), iface.(bool))
			case "string":
				val = iface.(string)
			case "[]string":
				if reflect.TypeOf(val) == nil {
					val = []string{}
				}
				val = reflect.Append(
					reflect.ValueOf(val),
					reflect.ValueOf(iface),
				).Interface()

			case "int":
				val = iface.(int)
			case "[]int":
				if reflect.TypeOf(val) == nil {
					val = []int{}
				}
				val = append(val.([]int), iface.(int))
			case "float64":
				val = iface.(float64)
			case "[]float64":
				if reflect.TypeOf(val) == nil {
					val = []float64{}
				}
				val = append(val.([]float64), iface.(float64))
			default:
				err = fmt.Errorf("Cannot set config from %v", value)
				fmt.Println("nope", err)
				return
			}

			err = c.setValue(key, val)
			if err != nil {
				break
			}
		}
		return
	}

	// fmt.Printf("set %s to a %v with value %v\n", key, c.ConfigTypes[key], value)
	// fmt.Println("-- Got", reflect.TypeOf(value))

	var is interface{}
	switch c.ConfigTypes[key] {
	case "bool":
		var v bool
		v, ok := value.(bool)
		if !ok {
			v, err = boolFromInterface(value)
			if err != nil {
				return
			}
		}
		field.SetBool(v)
	case "string":
		var v string
		v, ok := value.(string)
		if !ok {
			vs, ok := value.([]string)
			if ok {
				v = vs[0]
			}
		}
		field.SetString(v)
	case "int":
		var v int64
		v, ok := value.(int64)
		if !ok {
			v, err = intFromInterface(value)
			if err != nil {
				return
			}
		}
		field.SetInt(v)
	case "float64":
		var v float64
		v, ok := value.(float64)
		if !ok {
			v, err = floatFromInterface(value)
			if err != nil {
				return
			}
		}
		field.SetFloat(v)
	case "[]string", "*set.StringSet":
		// log.Println("[]string", "*set.StringSet")
		var v []string
		v, ok := value.([]string)
		if !ok {
			var type_ string
			is, err = sliceFromStrings(value, type_)
			if err != nil {
				return
			}
			v = is.([]string)
		}
		if c.ConfigTypes[key] == "*set.StringSet" {
			field.Set(reflect.ValueOf(set.NewStringSet(v...)))
		} else {
			field.Set(reflect.ValueOf(v))
		}

	case "[]int", "*set.IntSet":
		var v []int
		v, ok := value.([]int)
		if !ok {
			var type_ int
			is, err = sliceFromStrings(value, type_)
			if err != nil {
				return
			}
			v = is.([]int)
		}

		if c.ConfigTypes[key] == "*set.IntSet" {
			field.Set(reflect.ValueOf(set.NewIntSet(v...)))
		} else {
			field.Set(reflect.ValueOf(v))
		}
	case "[]float64":
		var v []float64
		v, ok := value.([]float64)
		if !ok {
			var type_ float64
			is, err = sliceFromStrings(value, type_)
			if err != nil {
				return
			}
			v = is.([]float64)
		}
		field.Set(reflect.ValueOf(v))
	case "interface{}":
		err = fmt.Errorf("Set interface{} - not implemented")

	case "[]interface{}":
		err = fmt.Errorf("Set []interface{} - not implemented")
	default:
		err = fmt.Errorf("Cannot deal with %v.", c.ConfigTypes[key])
	}

	if c.LogSet && err == nil {
		fld, _ := c.FieldForKey(key)
		log.Printf("Set %s to %v.\n", key, fld.Interface())
	}

	return
}

func (c *Config) setDefaults() (err error) {
	for key, tag := range c.ConfigTags {
		if tag.Get(TAG_DEFAULT) != "" {
			err = c.setValue(key, tag.Get(TAG_DEFAULT))
			if err != nil {
				return
			}
		}
	}
	return
}

func (c *Config) readFromEnv() (err error) {
	env := os.Environ()
	for _, enVar := range env {
		key, val := strings.Split(enVar, "=")[0], strings.Split(enVar, "=")[1]
		if len(key) >= len(ENV_PREFIX) && key[0:len(ENV_PREFIX)] == ENV_PREFIX &&
			c.ConfigKeys.HasMember(key[len(ENV_PREFIX):]) {
			log.Println(key[0:len(ENV_PREFIX)], key[len(ENV_PREFIX):])
			err = c.setValue(key[len(ENV_PREFIX):], val)
			if err != nil {
				log.Println("Error setting from ENV:", err)
				return
			}
		}
	}

	return
}

func (c *Config) setReloadable() (err error) {
	return
}

func (c *Config) extractAliases() (err error) {
	for key, tag := range c.ConfigTags {
		if tag.Get(TAG_ALIAS) != "" {
			if _, notSet := c.AliasKeyMap[tag.Get(TAG_ALIAS)]; !notSet {
				c.KeyAliases.Add(tag.Get(TAG_ALIAS))
				c.AliasKeyMap[tag.Get(TAG_ALIAS)] = key
				c.ConfigTypes[tag.Get(TAG_ALIAS)] = c.ConfigTypes[key]
			} else {
				err = fmt.Errorf("Alias %s already set for field %s. Aliases must be unique.", tag.Get(TAG_ALIAS), c.AliasKeyMap[tag.Get(TAG_ALIAS)])
				return
			}
		}
	}
	return
}

func (c *Config) extractAssertions() (err error) {
	for key, tag := range c.ConfigTags {
		if tag.Get(TAG_ASSERTION) != "" {
			for _, assertion := range strings.Split(tag.Get(TAG_ASSERTION), TAG_ASSERTION_SEP) {
				assertParts := strings.SplitN(assertion, TAG_ASSERTION_FIELD, 2)

				// check for scope change
				if len(assertParts) == 1 {
					assertParts = strings.SplitN(assertion, TAG_ASSERTION_VALUE, 2)
					if len(assertParts) > 1 {
						// log.Println("TAG_ASSERTION_VALUE", c.ConfigTypes[key], assertParts[1], assertParts)
					} else {
						assertParts = append(assertParts, key)
						// log.Println("one elm assert", assertParts, len(assertParts))
					}
				} else {
					if len(assertParts) > 1 &&
						!strings.Contains(assertParts[1], ".") {
						// use current scope
						if strings.Contains(key, ".") {
							pathParts := strings.SplitAfter(key, ".")
							assertParts[1] = fmt.Sprintf("%s%s", strings.Join(pathParts[:len(pathParts)-1], ""), assertParts[1])
						}
					} else {
						// leading `.` refers to the struct root level
						assertParts[1] = strings.TrimLeft(assertParts[1], ".")
					}

					// does the target field exist ?
					if !c.ConfigKeys.HasMember(assertParts[1]) {
						err = fmt.Errorf("Unknown field %s in assertion %s", assertParts[1], assertion)
						return
					}
				}

				// log.Println("assertParts", assertParts, len(assertParts))
				// does the assertion function exist?
				if _, notSet := configAssertions[assertParts[0]]; !notSet {
					err = fmt.Errorf("Unknown assertion function: %s", assertParts[0])
					return
				}

				if _, notSet := c.Assertions[key][assertParts[1]]; !notSet {
					c.Assertions[key][assertParts[1]] = []Assertion{}
				}
				c.Assertions[key][assertParts[1]] = append(c.Assertions[key][assertParts[1]], configAssertions[assertParts[0]])
			}
		}
		if (tag.Get(TAG_REQUIRED) != "" && strings.ToLower(tag.Get(TAG_REQUIRED)) == "true") ||
			tag.Get(TAG_DEFAULT) != "" {
			// log.Println("TAG_REQUIRED or TAG_DEFAULT for", key)
			c.Assertions[key][key] = append(c.Assertions[key][key], configAssertions["NotBeNil"])
			c.Assertions[key][key] = append(c.Assertions[key][key], configAssertions["NotBeEmpty"])
		}
	}
	return
}

func StructFields(iface interface{}) (names []string, types map[string]string, tags map[string]reflect.StructTag) {
	types = make(map[string]string)
	tags = make(map[string]reflect.StructTag)
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)

	for ift.Kind() == reflect.Ptr {
		ift = ift.Elem()
		ifv = ifv.Elem()
	}

	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		f := ift.Field(i)

		if f.Name == "Config" {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			n, typs, tgs := StructFields(v.Interface())
			names = append(names, func(nms []string) []string {
				fullNames := []string{}
				for _, n := range nms {
					fullNames = append(fullNames, fmt.Sprintf("%s.%s", f.Name, n))
				}
				return fullNames
			}(n)...)

			for key, typ := range typs {
				types[fmt.Sprintf("%s.%s", f.Name, key)] = typ
			}

			for key, tag := range tgs {
				tags[fmt.Sprintf("%s.%s", f.Name, key)] = tag
			}
		default:
			names = append(names, f.Name)
			types[f.Name] = v.Type().String()
			tags[f.Name] = f.Tag
			// if !f.Anonymous {
			// }
		}
	}

	return
}
