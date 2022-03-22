// Copyright(C) 2020 Baidu Inc. All Rights Reserved.
// Author: Chen Xin (chenxin@baidu.com)
// Date: 2020/04/19

package logit

import (
	"container/list"
	"context"
	"sync"
)

// logFieldKey 在context中使用的key
type logContextKey int

const (
	ctxLogFieldsKey logContextKey = iota
	ctxMetaFieldsKey
)

// WithContext 在当前ctx中预埋日志字段
// 若已进行预埋，将返回原ctx
//
// 一般情况下，在业务的handler中都是已经调用过了，不需要重复调用
func WithContext(ctx context.Context) context.Context {
	ctx = initMetaFields(ctx)

	if findLogFields(ctx) != nil {
		return ctx
	}
	return NewContext(ctx)
}

// ContextInitialized 判断当前 context 是否已经初始化过，可以存储日志字段了
func ContextInitialized(ctx context.Context) bool {
	return findLogFields(ctx) != nil
}

// initMetaFields 初始化meta field字段
// meta fields 是唯一的，不允许在ctx子节点出现子 meta fields
func initMetaFields(ctx context.Context) context.Context {
	if findMetaFields(ctx) == nil {
		ctx = context.WithValue(ctx, ctxMetaFieldsKey, newLogContextStructure())
	}
	return ctx
}

// NewContext 创建一个新的ctx，并预埋日志字段
func NewContext(ctx context.Context) context.Context {
	ctx = initMetaFields(ctx)
	return context.WithValue(ctx, ctxLogFieldsKey, newLogContextStructure())
}

// ForkContext 基于当前的 Context 复制出一个新的 Context 分支，新的 Context 继承原 Context 中的日志 field。
func ForkContext(ctx context.Context) context.Context {
	ctx = initMetaFields(ctx)
	if fields := findLogFields(ctx); fields != nil {
		return context.WithValue(ctx, ctxLogFieldsKey, fields.clone())
	}
	return WithContext(ctx)
}

// CopyAllFields copy all common fields and meta fields from src to dest
//
// 使用场景：
// 如在调用一个函数，期望在后台继续运行，使用了 context.Background() ，但是期望保持原有的日志字段(如 logid)
func CopyAllFields(dest, src context.Context) context.Context {
	ctx := WithContext(dest)
	Range(src, func(f Field) error {
		AddFields(ctx, f)
		return nil
	})
	RangeMetaFields(src, func(f Field) error {
		AddMetaFields(ctx, f)
		return nil
	})
	return ctx
}

// Range 遍历存储在ctx里的Fields
func Range(ctx context.Context, f func(f Field) error) {
	if fields := findLogFields(ctx); fields != nil {
		fields.rangeFields(f)
	}
}

func findLogFields(ctx context.Context) *logContextStructure {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(ctxLogFieldsKey); v != nil {
		if fm, ok := v.(*logContextStructure); ok {
			return fm
		}
	}
	return nil
}

func mustLogField(ctx context.Context) *logContextStructure {
	fields := findLogFields(ctx)
	if fields == nil {
		panic("context should use logit.WithContext to initial first")
	}
	return fields
}

// AddFields 批量添加field，如果Field没有设置Level的话，实际上并不生效。使用AddNotice, AddWarning
// 等方法，同时为Field设置Level，并保存到context中。
//
// 该方法不会对字段属性进行修改
func AddFields(ctx context.Context, fields ...Field) {
	mustLogField(ctx).addFields(fields...)
}

// ReplaceFields 批量替换fields
func ReplaceFields(ctx context.Context, fields ...Field) {
	mustLogField(ctx).replaceFields(fields...)
}

// FindField 查找Field,查找不到会返回nil
func FindField(ctx context.Context, key string) Field {
	fs := findLogFields(ctx)
	if fs == nil {
		return nil
	}
	return fs.findField(key)
}

// DeleteField 删除field
func DeleteField(ctx context.Context, keys ...string) {
	fs := findLogFields(ctx)
	if fs == nil {
		return
	}
	fs.delFields(keys...)
}

// AddDebug 添加Field，Debug级别可见，只有在调用Logger的Debug方法时字段才输出
//
// 被添加的字段会修改其 Level 属性，故一个Field不要重复使用
// 若需要复用，可以使用 AddFields 方法
func AddDebug(ctx context.Context, fields ...Field) {
	for i := range fields {
		fields[i].SetLevel(DebugLevel)
	}
	AddFields(ctx, fields...)
}

// AddTrace 添加Field，Trace级别可见，只有在调用Logger的Trace方法时字段才输出
//
// 被添加的字段会修改其 Level 属性，故一个Field不要重复使用
// 若需要复用，可以使用 AddFields 方法
func AddTrace(ctx context.Context, fields ...Field) {
	for i := range fields {
		fields[i].SetLevel(TraceLevel)
	}
	AddFields(ctx, fields...)
}

// AddNotice 添加Field，Notice级别可见，只有在调用Logger的Notice方法时字段才输出
//
// 被添加的字段会修改其 Level 属性，故一个Field不要重复使用
// 若需要复用，可以使用 AddFields 方法
func AddNotice(ctx context.Context, fields ...Field) {
	for i := range fields {
		fields[i].SetLevel(NoticeLevel)
	}
	AddFields(ctx, fields...)
}

// AddWarning 添加Field，Warning级别可见，只有在调用Logger的Warning方法时字段才输出
//
// 被添加的字段会修改其 Level 属性，故一个Field不要重复使用
// 若需要复用，可以使用 AddFields 方法
func AddWarning(ctx context.Context, fields ...Field) {
	for i := range fields {
		fields[i].SetLevel(WarningLevel)
	}
	AddFields(ctx, fields...)
}

// AddError 添加Field，Notice级别可见，只有在调用Logger的Error方法时字段才输出
//
// 被添加的字段会修改其 Level 属性，故一个Field不要重复使用
// 若需要复用，可以使用 AddFields 方法
func AddError(ctx context.Context, fields ...Field) {
	for i := range fields {
		fields[i].SetLevel(ErrorLevel)
	}
	AddFields(ctx, fields...)
}

// AddFatal 添加Field， Fatal级别可见，只有在调用Logger的Fatal方法时字段才输出
//
// 被添加的字段会修改其 Level 属性，故一个Field不要重复使用
// 若需要复用，可以使用 AddFields 方法
func AddFatal(ctx context.Context, fields ...Field) {
	for i := range fields {
		fields[i].SetLevel(FatalLevel)
	}
	AddFields(ctx, fields...)
}

// AddAllLevel 添加字段，所有等级可见，相当于php的 AddNotice 方法，调用Logger的任意方法，字段都将输出
//
// 被添加的字段会修改其 Level 属性，故一个Field不要重复使用
// 若需要复用，可以使用 AddFields 方法
func AddAllLevel(ctx context.Context, fields ...Field) {
	for i := range fields {
		fields[i].SetLevel(AllLevels)
	}
	AddFields(ctx, fields...)
}

func findMetaFields(ctx context.Context) *logContextStructure {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(ctxMetaFieldsKey); v != nil {
		if fm, ok := v.(*logContextStructure); ok {
			return fm
		}
	}
	return nil
}

func mustMetaField(ctx context.Context) *logContextStructure {
	fields := findMetaFields(ctx)
	if fields == nil {
		panic("context should use logit.WithContext to initial first")
	}
	return fields
}

// AddMetaFields 添加到全局 meta fields里，让字段在所有的日志里串联
//
// 	该方法添加的字段不会受到 NewContext 和  ForkContext的影响，创建新的ctx，之前添加的字段依然可见
// 	如 logid 字段,在不同的日志里，是需要保持串联的，就可以使用该方法传递
//
// 该方法不会对字段属性进行修改
func AddMetaFields(ctx context.Context, fields ...Field) {
	mustMetaField(ctx).addFields(fields...)
}

// FindMetaField 查找Field,查找不到会返回nil
func FindMetaField(ctx context.Context, key string) Field {
	fs := findMetaFields(ctx)
	if fs == nil {
		return nil
	}
	return fs.findField(key)
}

// DeleteMetaField 删除meta field
func DeleteMetaField(ctx context.Context, keys ...string) {
	fs := findMetaFields(ctx)
	if fs == nil {
		return
	}
	fs.delFields(keys...)
}

// ReplaceMetaFields 批量替换meta fields
func ReplaceMetaFields(ctx context.Context, fields ...Field) {
	mustMetaField(ctx).replaceFields(fields...)
}

// RangeMetaFields 遍历存储在ctx里的meta Fields
func RangeMetaFields(ctx context.Context, f func(f Field) error) {
	if fields := findMetaFields(ctx); fields != nil {
		fields.rangeFields(f)
	}
}

// logFieldInContext 在context中保存Field的数据结构
type logContextStructure struct {
	entry *list.List               // 按添加顺序存储的Field链表
	keys  map[string]*list.Element // field 的 key 与链表的映射
	mtx   sync.RWMutex
}

func newLogContextStructure() *logContextStructure {
	return &logContextStructure{
		entry: list.New(),
		keys:  make(map[string]*list.Element),
	}
}

func (lcs *logContextStructure) addFields(fs ...Field) {
	lcs.mtx.Lock()
	defer lcs.mtx.Unlock()

	for _, f := range fs {
		if old, ok := lcs.keys[f.Key()]; ok {
			lcs.entry.Remove(old)
		}
		lcs.keys[f.Key()] = lcs.entry.PushBack(f)
	}
}

func (lcs *logContextStructure) replaceFields(fs ...Field) {
	lcs.mtx.Lock()
	defer lcs.mtx.Unlock()

	for _, f := range fs {
		if old, ok := lcs.keys[f.Key()]; ok {
			if f.Level() == UnknownLevel {
				f.SetLevel(old.Value.(Field).Level())
			}
			lcs.keys[f.Key()].Value = f
		} else {
			// 不存在的时候，添加到尾部，同时若没有设置日志等级，则默认为 AllLevels
			if f.Level() == UnknownLevel {
				f.SetLevel(AllLevels)
			}
			lcs.keys[f.Key()] = lcs.entry.PushBack(f)
		}
	}
}

func (lcs *logContextStructure) delFields(keys ...string) {
	lcs.mtx.Lock()
	defer lcs.mtx.Unlock()

	for _, key := range keys {
		if f, ok := lcs.keys[key]; ok {
			lcs.entry.Remove(f)
			delete(lcs.keys, key)
		}
	}
}

func (lcs *logContextStructure) findField(key string) Field {
	lcs.mtx.RLock()
	defer lcs.mtx.RUnlock()

	if f, ok := lcs.keys[key]; ok {
		return f.Value.(Field)
	}
	return nil
}

func (lcs *logContextStructure) clone() *logContextStructure {
	lcs.mtx.RLock()
	defer lcs.mtx.RUnlock()

	copy := newLogContextStructure()
	for f := lcs.entry.Front(); f != nil; f = f.Next() {
		field := f.Value.(Field)
		copy.addFields(field)
	}
	return copy
}

func (lcs *logContextStructure) rangeFields(rangeFunc func(f Field) error) {
	lcs.mtx.RLock()
	defer lcs.mtx.RUnlock()

	for f := lcs.entry.Front(); f != nil; f = f.Next() {
		field := f.Value.(Field)
		if rangeFunc(field) != nil {
			break
		}
	}
}
