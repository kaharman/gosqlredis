// Copyright 2021 by Lukisno Kaharman. All rights reserved.
// This Source Code Form is subject to the terms of the Apache
// License 2.0 that can be found in the LICENSE file.

package gosqlredis

import (
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

type IndexType string

const (
	String IndexType = "string"
	Number IndexType = "number"
)

type Index struct {
	Name string
	Type IndexType
}

func GetListRedisDataToStructSorted(redisPool *redis.Pool, redisDatabase uint64, keyTable string, sqlStruct interface{}, desc bool, index Index, limit int, offset int) (int64, interface{}, error) {
	pool := redisPool.Get()
	defer pool.Close()
	pool.Do("SELECT", redisDatabase)

	total, err := redis.Int64(pool.Do("ZCARD", keyTable+":index:"+index.Name))
	if err != nil {
		return 0, nil, err
	}
	if total == 0 {
		return 0, nil, errors.New("no data")
	}

	var listID []int64
	if index.Type == Number {
		if desc {
			listID, err = redis.Int64s(pool.Do("ZRANGE", keyTable+":index:"+index.Name, MaxInt64, 0, "REV", "BYSCORE", "LIMIT", offset, limit))
		} else {
			listID, err = redis.Int64s(pool.Do("ZRANGE", keyTable+":index:"+index.Name, 0, MaxInt64, "BYSCORE", "LIMIT", offset, limit))
		}
		if err != nil {
			return 0, nil, err
		}
	} else {
		var listValueID []string
		if desc {
			listValueID, err = redis.Strings(pool.Do("ZRANGE", keyTable+":index:"+index.Name, "+", "-", "REV", "BYLEX", "LIMIT", offset, limit))
		} else {
			listValueID, err = redis.Strings(pool.Do("ZRANGE", keyTable+":index:"+index.Name, "-", "+", "BYLEX", "LIMIT", offset, limit))
		}
		if err != nil {
			return 0, nil, err
		}
		for _, valueID := range listValueID {
			valueID = valueID[strings.LastIndex(valueID, ":")+1:]
			id, err := strconv.ParseInt(valueID, 10, 64)
			if err == nil {
				listID = append(listID, id)
			}
		}
	}

	var result []interface{}
	for _, id := range listID {
		newSqlStruct := reflect.New(reflect.ValueOf(sqlStruct).Elem().Type()).Interface()
		sqlData, err := GetRedisDataToStruct(redisPool, redisDatabase, keyTable, newSqlStruct, strconv.FormatInt(id, 10))
		if err != nil {
			return 0, nil, err
		}
		rv := reflect.ValueOf(sqlData).Elem()
		result = append(result, rv.Interface())
	}

	return total, result, nil
}

func GetListRedisDataToStruct(redisPool *redis.Pool, redisDatabase uint64, keyTable string, sqlStruct interface{}) (int64, interface{}, error) {
	pool := redisPool.Get()
	defer pool.Close()
	pool.Do("SELECT", redisDatabase)

	total, err := redis.Int64(pool.Do("SCARD", keyTable+":index:id"))
	if err != nil {
		return 0, nil, err
	}
	if total == 0 {
		return 0, nil, errors.New("no data")
	}

	scan, err := redis.Values(pool.Do("SSCAN", keyTable+":index:id", 0, "COUNT", 10000000))
	if err != nil {
		return 0, nil, err
	}
	keys := scan[1]

	var result []interface{}
	for _, key := range keys.([]interface{}) {
		sqlData, err := GetRedisDataToStruct(redisPool, redisDatabase, keyTable, sqlStruct, string(key.([]byte)))
		if err != nil {
			return 0, nil, err
		}
		rv := reflect.ValueOf(sqlData).Elem()
		result = append(result, rv.Interface())
	}

	return total, result, nil
}

func GetRedisDataToStruct(redisPool *redis.Pool, redisDatabase uint64, keyTable string, sqlStruct interface{}, id string) (interface{}, error) {
	pool := redisPool.Get()
	defer pool.Close()
	pool.Do("SELECT", redisDatabase)

	values, err := redis.Values(pool.Do("HGETALL", keyTable+":"+id))
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, errors.New("no data")
	}

	sqlData := sqlStruct
	rt := reflect.TypeOf(sqlData).Elem()
	rv := reflect.ValueOf(sqlData).Elem()
	for i := 0; i < rt.NumField(); i++ {
		structField := rt.Field(i)
		for j := 0; j < len(values); j = j + 2 {
			if structField.Tag.Get("redis") == string(values[j].([]byte)) {
				switch structField.Type.Kind().String() {
				case "string":
					rv.Field(i).SetString(string(values[j+1].([]byte)))
				case "bool":
					val, err := strconv.ParseBool(string(values[j+1].([]byte)))
					if err == nil {
						rv.Field(i).SetBool(val)
					}
				case "int", "int8", "int16", "int32", "int64":
					val, err := strconv.ParseInt(string(values[j+1].([]byte)), 10, 64)
					if err == nil {
						rv.Field(i).SetInt(val)
					}
				case "uint", "uint8", "uint16", "uint32", "uint64":
					val, err := strconv.ParseUint(string(values[j+1].([]byte)), 10, 64)
					if err == nil {
						rv.Field(i).SetUint(val)
					}
				case "float32", "float64":
					val, err := strconv.ParseFloat(string(values[j+1].([]byte)), 64)
					if err == nil {
						rv.Field(i).SetFloat(val)
					}
				case "struct":
					if structField.Type.PkgPath() == "database/sql" {
						switch structField.Type.Name() {
						case "NullBool":
							val, err := strconv.ParseBool(string(values[j+1].([]byte)))
							if err == nil {
								rv.Field(i).FieldByName("Bool").SetBool(val)
								rv.Field(i).FieldByName("Valid").SetBool(true)
							}
						case "NullFloat64":
							val, err := strconv.ParseFloat(string(values[j+1].([]byte)), 64)
							if err == nil {
								rv.Field(i).FieldByName("Float64").SetFloat(val)
								rv.Field(i).FieldByName("Valid").SetBool(true)
							}
						case "NullInt32":
							val, err := strconv.ParseInt(string(values[j+1].([]byte)), 10, 64)
							if err == nil {
								rv.Field(i).FieldByName("Int32").SetInt(val)
								rv.Field(i).FieldByName("Valid").SetBool(true)
							}
						case "NullInt64":
							val, err := strconv.ParseInt(string(values[j+1].([]byte)), 10, 64)
							if err == nil {
								rv.Field(i).FieldByName("Int64").SetInt(val)
								rv.Field(i).FieldByName("Valid").SetBool(true)
							}
						case "NullString":
							rv.Field(i).FieldByName("String").SetString(string(values[j+1].([]byte)))
							rv.Field(i).FieldByName("Valid").SetBool(true)
						case "NullTime":
							dateTime, err := time.Parse(time.RFC3339, string(values[j+1].([]byte)))
							if err == nil {
								rv.Field(i).FieldByName("Time").Set(reflect.ValueOf(dateTime))
								rv.Field(i).FieldByName("Valid").SetBool(true)
							}
						}
					} else if structField.Type.PkgPath() == "time" {
						if structField.Type.Name() == "Time" {
							dateTime, err := time.Parse(time.RFC3339, string(values[j+1].([]byte)))
							if err == nil {
								rv.Field(i).Set(reflect.ValueOf(dateTime))
							}
						}
					}
				}
			}
		}
	}

	return sqlData, nil
}

func InsertSQLDataToRedis(redisPool *redis.Pool, sqlData interface{}, redisDatabase uint64, keyTable string, id string, index ...Index) error {
	pool := redisPool.Get()
	defer pool.Close()
	pool.Do("SELECT", redisDatabase)

	var params []interface{}
	params = append(params, keyTable+":"+id)
	rt := reflect.TypeOf(sqlData).Elem()
	rv := reflect.ValueOf(sqlData).Elem()
	var indexCommand [][]interface{}
	for i := 0; i < rt.NumField(); i++ {
		structField := rt.Field(i)
		value := rv.Field(i).Interface()
		var idx Index
		for _, indexVal := range index {
			if indexVal.Name == structField.Tag.Get("redis") {
				idx = indexVal
			}
		}
		if structField.Type.PkgPath() == "database/sql" {
			switch structField.Type.Name() {
			case "NullBool":
				if value.(sql.NullBool).Valid {
					params = append(params, structField.Tag.Get("redis"), value.(sql.NullBool).Bool)
				}
			case "NullFloat64":
				if value.(sql.NullFloat64).Valid {
					params = append(params, structField.Tag.Get("redis"), value.(sql.NullFloat64).Float64)
				}
			case "NullInt32":
				if value.(sql.NullInt32).Valid {
					params = append(params, structField.Tag.Get("redis"), value.(sql.NullInt32).Int32)
				}
			case "NullInt64":
				if value.(sql.NullInt64).Valid {
					params = append(params, structField.Tag.Get("redis"), value.(sql.NullInt64).Int64)
				}
			case "NullString":
				if value.(sql.NullString).Valid {
					params = append(params, structField.Tag.Get("redis"), value.(sql.NullString).String)
				}
			case "NullTime":
				if value.(sql.NullTime).Valid {
					params = append(params, structField.Tag.Get("redis"), value.(sql.NullTime).Time.Format(time.RFC3339))
				}
			}
		} else if structField.Type.PkgPath() == "time" {
			params = append(params, structField.Tag.Get("redis"), value.(time.Time).Format(time.RFC3339))
			if idx.Type == String {
				var indexParams []interface{}
				indexParams = append(indexParams, keyTable+":index:"+idx.Name, 0, value.(time.Time).Format(time.RFC3339)+":"+id)
				indexCommand = append(indexCommand, indexParams)
			}
		} else if structField.Type.Kind().String() == "string" || structField.Type.Kind().String() == "struct" {
			params = append(params, structField.Tag.Get("redis"), value)
			if idx.Type == String {
				var indexParams []interface{}
				indexParams = append(indexParams, keyTable+":index:"+idx.Name, 0, value.(string)+":"+id)
				indexCommand = append(indexCommand, indexParams)
			}
		} else {
			params = append(params, structField.Tag.Get("redis"), value)
			if idx.Type == Number {
				var indexParams []interface{}
				indexParams = append(indexParams, keyTable+":index:"+idx.Name, value, id)
				indexCommand = append(indexCommand, indexParams)
			}
		}
	}

	_, err := pool.Do("HSET", params...)
	if err != nil {
		return err
	}
	pool.Do("SADD", keyTable+":index:id", id)
	for _, indexParams := range indexCommand {
		pool.Do("ZADD", indexParams...)
	}

	return nil
}

func DeleteDataRedis(redisPool *redis.Pool, redisDatabase uint64, keyTable string, id string, index ...Index) error {
	pool := redisPool.Get()
	defer pool.Close()
	pool.Do("SELECT", redisDatabase)

	_, err := pool.Do("SREM", keyTable+":index:id", id)
	if err != nil {
		return err
	}
	for _, idx := range index {
		var member string
		if idx.Type == IndexType(String) {
			scan, err := redis.Values(pool.Do("ZSCAN", keyTable+":index:"+idx.Name, 0, "MATCH", "*:"+id, "COUNT", 10000000))
			if err != nil {
				return err
			}
			keys := scan[1]
			members := keys.([]interface{})
			if len(members) == 0 {
				continue
			}
			member = string(members[0].([]byte))
		} else {
			_, err := redis.String(pool.Do("ZSCORE", keyTable+":index:"+idx.Name, id))
			if err != nil {
				continue
			}
			member = id
		}
		_, err := pool.Do("ZREM", keyTable+":index:"+idx.Name, member)
		if err != nil {
			return err
		}
	}
	_, err = pool.Do("DEL", keyTable+":"+id)
	if err != nil {
		return err
	}

	return nil
}

func ClearDataRedis(redisPool *redis.Pool, redisDatabase uint64, keyTable string) error {
	pool := redisPool.Get()
	defer pool.Close()
	pool.Do("SELECT", redisDatabase)

	scan, err := redis.Values(pool.Do("SCAN", 0, "MATCH", keyTable+":*", "COUNT", 10000000))
	if err != nil {
		return err
	}
	keys := scan[1]
	if len(keys.([]interface{})) == 0 {
		return errors.New("no data in redis")
	}

	for _, key := range keys.([]interface{}) {
		_, err = pool.Do("DEL", string(key.([]byte)))
		if err != nil {
			return err
		}
	}

	return nil
}
