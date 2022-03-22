#!/bin/env python3
# -*- coding: utf-8 -*-

""" generate code pieces
logit.Field has many primitive types, this script can generate codes for all types with template
"""

import sys

fieldTypes = [
    ["Binary", "[]byte"],
    ["Bool", "bool"],
    ["ByteString", "[]byte"],
    ["Duration", "time.Duration"],
    ["Float64", "float64"],
    ["Float32", "float32"],
    ["Int", "int"],
    ["Int64", "int64"],
    ["Int32", "int32"],
    ["Int16", "int16"],
    ["Int8", "int8"],
    ["String", "string"],
    ["Time", "time.Time"],
    ["Uint", "uint"],
    ["Uint64", "uint64"],
    ["Uint32", "uint32"],
    ["Uint16", "uint16"],
    ["Uint8", "uint8"],
    ["Uintptr", "uintptr"],
    ["Reflect", "interface{}"],
    ["Error", "error"],
]

# fieldCreator 生成 func(key string, value PRIMITIVE_TYPE) Field 的方法
fieldCreateTPL = """// {TYPE_NAME} field creator
func {TYPE_NAME}(key string, value {PRIMITIVE_TYPE}) Field {
    return &field{
        fieldType: {TYPE_NAME}Type,
        key: key,
        value: value,
    }
}"""
def fieldCreator():
    """ create fielCreator with type"""
    for type_name, prim_type in fieldTypes:
        s = fieldCreateTPL.replace("{TYPE_NAME}", type_name)
        s = s.replace("{PRIMITIVE_TYPE}", prim_type)
        print(s)

# addTo 生成 Field.AddTo(enc Encoder) 的 case 列表
addToTPL = """case {TYPE_NAME}Type:
    enc.Add{TYPE_NAME}(f.Key, f.Value.({PRIMITIVE_TYPE}))"""
def addTo():
    """ create addTo cases with type"""
    for type_name, prim_type in fieldTypes:
        s = addToTPL.replace("{TYPE_NAME}", type_name)
        s = s.replace("{PRIMITIVE_TYPE}", prim_type)
        print(s)

# autoField 生成 AutoField 中的 case 列表
autoFieldTPL = """case {PRIMITIVE_TYPE}:
    return {TYPE_NAME}(key, val)"""
def autoField():
    """ create autoField cases with type"""
    for type_name, prim_type in fieldTypes:
        s = autoFieldTPL.replace("{TYPE_NAME}", type_name)
        s = s.replace("{PRIMITIVE_TYPE}", prim_type)
        print(s)

# jsonEncoder 生成 JSONFieldEncoder 的接口方法
jsonEncoderTPL = """func (e *JSONFieldEncoder) Add{TYPE_NAME}(key string, value {PRIMITIVE_TYPE}) {
    e.kv[key] = value
}
"""
def jsonEncoder():
    """create json encoder for JSONFieldEncoder"""
    for type_name, prim_type in fieldTypes:
        s = jsonEncoderTPL.replace("{TYPE_NAME}", type_name)
        s = s.replace("{PRIMITIVE_TYPE}", prim_type)
        print(s)

def usage():
    """Display usage"""
    print("%s [field-creator|add-to|auto-field|json-encoder]" % (sys.argv[0]))

if __name__ == "__main__":
    if len(sys.argv) != 2:
        usage()
        exit(0)
    if sys.argv[1] == "field-creator":
        fieldCreator()
    elif sys.argv[1] == "add-to":
        addTo()
    elif sys.argv[1] == "auto-field":
        autoField()
    elif sys.argv[1] == "json-encoder":
        jsonEncoder()
    else:
        usage()