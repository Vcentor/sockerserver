// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Chen Xin (chenxin@baidu.com)
// Date: 2020/04/19

package logit

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// FieldType 表明当前 field 的类型
type FieldType uint8

const (
	// UnknownType is the default field type. Attempting to add it to an encoder will panic.
	UnknownType FieldType = iota

	// BinaryType indicates that the field carries an arbitrary binary data.
	BinaryType

	// BoolType indicates that the field carries a bool.
	BoolType

	// ByteStringType indicates that the field carries UTF-8 encoded bytes.
	ByteStringType

	// DurationType indicates that the field carries a time.duration.
	DurationType

	// Float64Type indicates that the field carries a float64.
	Float64Type

	// Float32Type indicates that the field carries a float32.
	Float32Type

	// IntType indicates that the field carries an int
	IntType

	// Int64Type indicates that the field carries an int64.
	Int64Type

	// Int32Type indicates that the field carries an int32.
	Int32Type

	// Int16Type indicates that the field carries an int16.
	Int16Type

	// Int8Type indicates that the field carries an int8.
	Int8Type

	// StringType indicates that the field carries a string.
	StringType

	// TimeType indicates that the field carries a time.Time that is
	// representable by a UnixNano() stored as an int64.
	TimeType

	// UintType indicates that the field carries an int
	UintType

	// Uint64Type indicates that the field carries a uint64.
	Uint64Type

	// Uint32Type indicates that the field carries a uint32.
	Uint32Type

	// Uint16Type indicates that the field carries a uint16.
	Uint16Type

	// Uint8Type indicates that the field carries a uint8.
	Uint8Type

	// UintptrType indicates that the field carries a uintptr.
	UintptrType

	// ReflectType indicates that the field carries an interface{}, which should
	// be serialized using reflection.
	ReflectType

	// ErrorType indicates that the field carries an error.
	ErrorType

	// DeferType 延迟获取值的类型
	DeferType
)

// Field 一个日志字段
type Field interface {
	// 字段名字
	Key() string

	// 字段值类型
	Type() FieldType

	// 字段值
	Value() interface{}

	// 字段的日志等级
	Level() Level

	// 修改日志等级
	// 其他的key，type，value是不允许修改的
	// 如 value修改，则type也需要更正变化，会引发已系列的问题
	// 若要修改type、value字段，应该创建一个新的field，然后使用ReplaceFields方法Replace掉
	SetLevel(level Level)

	// 是否相等
	// 不会比对Level
	Equal(other Field) bool

	AddTo(enc FieldEncoder)
}

// A field is a marshaling operation used to add a key-value pair to a SimpleLogger's
// context. Most fields are lazily marshaled, so it's inexpensive to add fields
// to disabled debug-level log statements.
type field struct {
	// Type field 的类型
	fieldType FieldType
	// level field 的日志等级
	level Level
	// key field 的 key
	key string
	// Value field 的 Value
	value interface{}
}

func (f *field) Type() FieldType {
	return f.fieldType
}

func (f *field) Level() Level {
	return f.level
}

func (f *field) SetLevel(level Level) {
	f.level = level
}

func (f *field) Key() string {
	return f.key
}

func (f *field) Value() interface{} {
	return f.value
}

// AddTo exports a field through the ObjectEncoder interface. It's primarily
// useful to library authors, and shouldn't be necessary in most applications.
func (f *field) AddTo(enc FieldEncoder) {
	FieldAddToEncoder(f, enc)
}

// FieldAddToEncoder 将字段添加到encoder
func FieldAddToEncoder(f Field, enc FieldEncoder) {
	var err error

	switch f.Type() {
	case BinaryType:
		enc.AddBinary(f.Key(), f.Value().([]byte))
	case BoolType:
		enc.AddBool(f.Key(), f.Value().(bool))
	case ByteStringType:
		enc.AddByteString(f.Key(), f.Value().([]byte))
	case DurationType:
		enc.AddDuration(f.Key(), f.Value().(time.Duration))
	case Float64Type:
		enc.AddFloat64(f.Key(), f.Value().(float64))
	case Float32Type:
		enc.AddFloat32(f.Key(), f.Value().(float32))
	case IntType:
		enc.AddInt(f.Key(), f.Value().(int))
	case Int64Type:
		enc.AddInt64(f.Key(), f.Value().(int64))
	case Int32Type:
		enc.AddInt32(f.Key(), f.Value().(int32))
	case Int16Type:
		enc.AddInt16(f.Key(), f.Value().(int16))
	case Int8Type:
		enc.AddInt8(f.Key(), f.Value().(int8))
	case StringType:
		enc.AddString(f.Key(), f.Value().(string))
	case TimeType:
		enc.AddTime(f.Key(), f.Value().(time.Time))
	case UintType:
		enc.AddUint(f.Key(), f.Value().(uint))
	case Uint64Type:
		enc.AddUint64(f.Key(), f.Value().(uint64))
	case Uint32Type:
		enc.AddUint32(f.Key(), f.Value().(uint32))
	case Uint16Type:
		enc.AddUint16(f.Key(), f.Value().(uint16))
	case Uint8Type:
		enc.AddUint8(f.Key(), f.Value().(uint8))
	case UintptrType:
		enc.AddUintptr(f.Key(), f.Value().(uintptr))
	case ReflectType:
		err = enc.AddReflected(f.Key(), f.Value())
	case ErrorType:
		if value, ok := f.Value().(error); ok {
			enc.AddError(f.Key(), value)
		} else {
			enc.AddError(f.Key(), nil)
		}
	case DeferType:
		fn := f.Value().(func() interface{})
		AutoField(f.Key(), fn()).AddTo(enc)
	default:
		panic(fmt.Sprintf("unknown field type: %v", f))
	}

	if err != nil {
		enc.AddString(fmt.Sprintf("%sError", f.Key()), err.Error())
	}
}

// Equals returns whether two fields are equal. For non-primitive types such as
// errors, or reflect types, it uses reflect.DeepEqual.
func (f *field) Equal(other Field) bool {
	return FieldEqual(f, other)
}

// FieldEqual 比较两个字段是否相等
// 不会比对Level
func FieldEqual(a, b Field) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if a.Type() != b.Type() {
		return false
	}

	if a.Key() != b.Key() {
		return false
	}

	switch a.Type() {
	case BinaryType, ByteStringType:
		return bytes.Equal(a.Value().([]byte), b.Value().([]byte))
	case ErrorType, ReflectType:
		return reflect.DeepEqual(a, b)
	default:
		return a.Value() == b.Value()
	}
}

var _ Field = (*field)(nil)

// Binary field creator
func Binary(key string, value []byte) Field {
	return &field{
		fieldType: BinaryType,
		key:       key,
		value:     value,
	}
}

// Bool field creator
func Bool(key string, value bool) Field {
	return &field{
		fieldType: BoolType,
		key:       key,
		value:     value,
	}
}

// ByteString field creator
func ByteString(key string, value []byte) Field {
	return &field{
		fieldType: ByteStringType,
		key:       key,
		value:     value,
	}
}

// Duration 时长字段
func Duration(key string, value time.Duration) Field {
	return &field{
		fieldType: DurationType,
		key:       key,
		value:     value,
	}
}

// Float64 field creator
func Float64(key string, value float64) Field {
	return &field{
		fieldType: Float64Type,
		key:       key,
		value:     value,
	}
}

// Float32 field creator
func Float32(key string, value float32) Field {
	return &field{
		fieldType: Float32Type,
		key:       key,
		value:     value,
	}
}

// Int field creator
func Int(key string, value int) Field {
	return &field{
		fieldType: IntType,
		key:       key,
		value:     value,
	}
}

// Int64 field creator
func Int64(key string, value int64) Field {
	return &field{
		fieldType: Int64Type,
		key:       key,
		value:     value,
	}
}

// Int32 field creator
func Int32(key string, value int32) Field {
	return &field{
		fieldType: Int32Type,
		key:       key,
		value:     value,
	}
}

// Int16 field creator
func Int16(key string, value int16) Field {
	return &field{
		fieldType: Int16Type,
		key:       key,
		value:     value,
	}
}

// Int8 field creator
func Int8(key string, value int8) Field {
	return &field{
		fieldType: Int8Type,
		key:       key,
		value:     value,
	}
}

// String field creator
func String(key string, value string) Field {
	return &field{
		fieldType: StringType,
		key:       key,
		value:     value,
	}
}

// Time field creator
func Time(key string, value time.Time) Field {
	return &field{
		fieldType: TimeType,
		key:       key,
		value:     value,
	}
}

// Uint field creator
func Uint(key string, value uint) Field {
	return &field{
		fieldType: UintType,
		key:       key,
		value:     value,
	}
}

// Uint64 field creator
func Uint64(key string, value uint64) Field {
	return &field{
		fieldType: Uint64Type,
		key:       key,
		value:     value,
	}
}

// Uint32 field creator
func Uint32(key string, value uint32) Field {
	return &field{
		fieldType: Uint32Type,
		key:       key,
		value:     value,
	}
}

// Uint16 field creator
func Uint16(key string, value uint16) Field {
	return &field{
		fieldType: Uint16Type,
		key:       key,
		value:     value,
	}
}

// Uint8 field creator
func Uint8(key string, value uint8) Field {
	return &field{
		fieldType: Uint8Type,
		key:       key,
		value:     value,
	}
}

// Uintptr field creator
func Uintptr(key string, value uintptr) Field {
	return &field{
		fieldType: UintptrType,
		key:       key,
		value:     value,
	}
}

// Reflect field creator
func Reflect(key string, value interface{}) Field {
	return &field{
		fieldType: ReflectType,
		key:       key,
		value:     value,
	}
}

// Error field creator
func Error(key string, value error) Field {
	return &field{
		fieldType: ErrorType,
		key:       key,
		value:     value,
	}
}

// Defer field creator
func Defer(key string, value func() interface{}) Field {
	return &field{
		fieldType: DeferType,
		key:       key,
		value:     value,
	}
}

// AutoField field creator
func AutoField(key string, value interface{}) Field {
	switch val := value.(type) {
	case []byte:
		return Binary(key, val)
	case bool:
		return Bool(key, val)
	case time.Duration:
		return Duration(key, val)
	case float64:
		return Float64(key, val)
	case float32:
		return Float32(key, val)
	case int:
		return Int(key, val)
	case int64:
		return Int64(key, val)
	case int32:
		return Int32(key, val)
	case int16:
		return Int16(key, val)
	case int8:
		return Int8(key, val)
	case string:
		return String(key, val)
	case time.Time:
		return Time(key, val)
	case uint:
		return Uint(key, val)
	case uint64:
		return Uint64(key, val)
	case uint32:
		return Uint32(key, val)
	case uint16:
		return Uint16(key, val)
	case uint8:
		return Uint8(key, val)
	case uintptr:
		return Uintptr(key, val)
	case error:
		return Error(key, val)
	case func() interface{}:
		return Defer(key, val)
	}
	return Reflect(key, value)
}

// KVs2fields key-value slice to fields slice
func KVs2fields(kvs ...interface{}) []Field {
	if len(kvs) < 2 {
		return nil
	}
	fields := make([]Field, 0, len(kvs)/2)
	key := true
	var keyName string
	for _, kv := range kvs {
		if key {
			var ok bool
			if keyName, ok = kv.(string); ok {
				key = false
			}
			continue
		}
		fields = append(fields, AutoField(keyName, kv))
		key = true
	}
	return fields
}

// FieldLogID 字段名，logid
const FieldLogID = "logid"

// LogIDField logid 字段
func LogIDField(logid string) Field {
	return String(FieldLogID, logid)
}

// FindLogIDField 查找logid字段
func FindLogIDField(ctx context.Context) Field {
	return FindMetaField(ctx, FieldLogID)
}

// SetLogID 将logid 设置到ctx里,logid 可以是string 和 整数类型
func SetLogID(ctx context.Context, logid interface{}) {
	var field Field
	var logidStr string
	switch val := logid.(type) {
	case Field:
		field = val
	case string:
		logidStr = val
	case int:
		logidStr = strconv.Itoa(val)
	case int8:
		logidStr = strconv.Itoa(int(val))
	case int16:
		logidStr = strconv.Itoa(int(val))
	case int32:
		logidStr = strconv.FormatInt(int64(val), 10)
	case int64:
		logidStr = strconv.FormatInt(val, 10)
	case uint:
		logidStr = strconv.FormatUint(uint64(val), 10)
	case uint8:
		logidStr = strconv.FormatUint(uint64(val), 10)
	case uint16:
		logidStr = strconv.FormatUint(uint64(val), 10)
	case uint32:
		logidStr = strconv.FormatUint(uint64(val), 10)
	case uint64:
		logidStr = strconv.FormatUint(val, 10)
	default:
		logidStr = NewLogID()
	}

	if logidStr != "" {
		field = LogIDField(logidStr)
	}
	ReplaceMetaFields(ctx, field)
	ReplaceFields(ctx, field)
}
