package hocon

import (
	"bytes"
	"fmt"
	"strings"
)

type HoconObject struct {
	items map[string]*HoconValue
	keys  []string
}

func NewHoconObject() *HoconObject {
	return &HoconObject{
		items: make(map[string]*HoconValue),
	}
}

func (p *HoconObject) GetString() string {
	panic("This element is an object and not a string.")
}

func (p *HoconObject) IsArray() bool {
	return false
}

func (p *HoconObject) GetArray() []*HoconValue {
	panic("This element is an object and not an array.")
}

func (p *HoconObject) GetKeys() []string {
	return p.keys
}

func (p *HoconObject) Unwrapped() map[string]interface{} {
	if len(p.items) == 0 {
		return nil
	}

	dics := map[string]interface{}{}

	for _, k := range p.keys {
		v := p.items[k]

		obj := v.GetObject()
		if obj != nil {
			dics[k] = obj.Unwrapped()
		} else {
			dics[k] = v
		}
	}

	return dics
}

func (p *HoconObject) Items() map[string]*HoconValue {
	return p.items
}

func (p *HoconObject) GetKey(key string) *HoconValue {
	value, _ := p.items[key]
	return value
}

func (p *HoconObject) GetOrCreateKey(key string) *HoconValue {
	if value, exist := p.items[key]; exist {
		child := NewHoconValue()
		child.oldValue = value
		p.items[key] = child
		return child
	}

	child := NewHoconValue()
	p.items[key] = child
	p.keys = append(p.keys, key)
	return child
}

func (p *HoconObject) IsString() bool {
	return false
}

func (p *HoconObject) String() string {
	return p.ToString(0)
}

func (p *HoconObject) ToString(indent int) string {
	tmp := strings.Repeat(" ", indent*2)
	buf := bytes.NewBuffer(nil)
	for _, k := range p.keys {
		key := p.quoteIfNeeded(k)
		v := p.items[key]
		buf.WriteString(fmt.Sprintf("%s%s : %s\r\n", tmp, key, v.ToString(indent)))
	}
	return buf.String()
}

func (p *HoconObject) Merge(other *HoconObject) {
	thisValues := p.items
	otherItems := other.items

	otherKeys := other.keys

	for _, otherkey := range otherKeys {
		otherValue := otherItems[otherkey]

		if thisValue, exist := thisValues[otherkey]; exist {
			if thisValue.IsObject() && otherValue.IsObject() {
				thisValue.GetObject().Merge(otherValue.GetObject())
			}
		} else {
			p.items[otherkey] = otherValue
			p.keys = append(p.keys, otherkey)
		}
	}
}

func (p *HoconObject) MergeImmutable(other *HoconObject) *HoconObject {
	thisValues := make(map[string]*HoconValue)
	thisKeys := make([]string, 0, len(p.keys))
	resource := p.items
	otherKeys := other.keys
	otherItems := other.items

	for _, otherkey := range otherKeys {
		otherValue := otherItems[otherkey]

		if thisValue, exist := resource[otherkey]; exist {

			if thisValue.IsObject() && otherValue.IsObject() {

				mergedObject := thisValue.GetObject().MergeImmutable(otherValue.GetObject())
				mergedValue := NewHoconValue()

				mergedValue.AppendValue(mergedObject)
				thisValues[otherkey] = mergedValue
			}
		} else {
			thisValues[otherkey] = &HoconValue{values: otherValue.values}
			thisKeys = append(thisKeys, otherkey)
		}
	}

	return &HoconObject{items: thisValues, keys: thisKeys}
}

func (p *HoconObject) Combine(other *HoconObject) {
	thisValues := p.items
	otherItems := other.items

	otherKeys := other.keys

	for _, otherKey := range otherKeys {
		otherValue := otherItems[otherKey]

		if thisValue, exist := thisValues[otherKey]; exist {
			if thisValue.IsObject() && otherValue.IsObject() {
				thisValue.GetObject().Combine(otherValue.GetObject())
			} else if otherValue.values != nil {
				thisValue.values = otherValue.values
			}
		} else {
			p.items[otherKey] = otherValue
			p.keys = append(p.keys, otherKey)
		}
	}
}

func (p *HoconObject) CombineImmutable(other *HoconObject) *HoconObject {
	thisValues := make(map[string]*HoconValue)
	thisKeys := make([]string, 0, len(p.keys))
	resource := p.items
	otherKeys := other.keys
	otherItems := other.items

	for _, otherKey := range otherKeys {
		otherValue := otherItems[otherKey]

		if thisValue, exist := resource[otherKey]; exist {

			if thisValue.IsObject() && otherValue.IsObject() {

				mergedObject := thisValue.GetObject().CombineImmutable(otherValue.GetObject())
				mergedValue := NewHoconValue()

				mergedValue.AppendValue(mergedObject)
				thisValues[otherKey] = mergedValue
			} else if otherValue.values != nil {
				thisValues[otherKey] = &HoconValue{values: otherValue.values}
			} else {
				thisValues[otherKey] = &HoconValue{values: thisValue.values}
			}
		} else {
			thisValues[otherKey] = &HoconValue{values: otherValue.values}
			thisKeys = append(thisKeys, otherKey)
		}
	}
	return &HoconObject{items: thisValues, keys: thisKeys}
}

func (p *HoconObject) quoteIfNeeded(text string) string {
	if strings.IndexByte(text, ' ') >= 0 ||
		strings.IndexByte(text, '\t') >= 0 {
		return "\"" + text + "\""
	}
	return text
}
