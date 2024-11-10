package hyaml

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"reflect"
)

func Unmarshal(in []byte, out interface{}) (err error) {
	return yaml.Unmarshal(in, out)
}

func MustUnmarshal(in []byte, out interface{}) {
	err := Unmarshal(in, out)
	if err != nil {
		panic(err)
	}
}

func UnmarshalStr(in string, out interface{}) (err error) {
	return Unmarshal([]byte(in), out)
}

func MustUnmarshalStr(in string, out interface{}) {
	MustUnmarshal([]byte(in), out)
}

func Marshal(in interface{}) (out []byte, err error) {
	return yaml.Marshal(in)
}

func MustMarshal(in interface{}) (out []byte) {
	bytes, err := Marshal(in)
	if err != nil {
		panic(err)
	}
	return bytes
}

func MarshalStr(in interface{}) (out string, err error) {
	bytes, err := Marshal(in)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func MustMarshalStr(in interface{}) (out string) {
	return string(MustMarshal(in))
}

func MarshalWithComment(in interface{}) (out []byte, err error) {
	var node yaml.Node
	err = node.Encode(in)
	if err != nil {
		return nil, err
	}
	yamlWithComment(&node, reflect.ValueOf(in), reflect.TypeOf(in))
	return Marshal(node)
}

func MarshalWithCommentT(in interface{}) (out []byte, err error) {
	var node yaml.Node
	err = node.Encode(in)
	if err != nil {
		return nil, err
	}
	yamlWithCommentAndType(&node, reflect.ValueOf(in), reflect.TypeOf(in))
	return Marshal(node)
}

func MarshalWithCommentStr(in interface{}) (out string, err error) {
	bytes, err := MarshalWithComment(in)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func MarshalWithCommentStrT(in interface{}) (out string, err error) {
	bytes, err := MarshalWithCommentT(in)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func MustMarshalWithComment(in interface{}) (out []byte) {
	bytes, err := MarshalWithComment(in)
	if err != nil {
		panic(err)
	}
	return bytes
}

func MustMarshalWithCommentStr(in interface{}) (out string) {
	return string(MustMarshalWithComment(in))
}

func yamlWithComment(node *yaml.Node, rv reflect.Value, rt reflect.Type) {
	switch rv.Kind() {
	case reflect.Pointer:
		yamlWithComment(node, rv.Elem(), rt.Elem())
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			tf := rt.Field(i)
			tag := tf.Tag.Get("yaml")
			if tag == "-" {
				continue
			}
			fieldTag := tf.Tag.Get("comment")
			hfieldTag := tf.Tag.Get("hcomment")
			ffieldTag := tf.Tag.Get("fcomment")
			node.Content[i*2].LineComment = fieldTag
			node.Content[i*2+1].LineComment = fieldTag
			node.Content[i*2].HeadComment = hfieldTag
			//node.Content[i*2+1].HeadComment = hfieldTag
			node.Content[i*2].FootComment = ffieldTag
			//node.Content[i*2+1].FootComment = ffieldTag
			yamlWithComment(node.Content[i*2+1], rv.Field(i), tf.Type)
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			index := rv.Index(i)
			yamlWithComment(node.Content[i], index, index.Type())
		}
	case reflect.Map:
		for i, key := range rv.MapKeys() {
			value := rv.MapIndex(key)
			valueType := value.Type()
			yamlWithComment(node.Content[i*2+1], value, valueType)
		}
	}
}

func yamlWithCommentAndType(node *yaml.Node, rv reflect.Value, rt reflect.Type) {
	switch rv.Kind() {
	case reflect.Pointer:
		yamlWithCommentAndType(node, rv.Elem(), rt.Elem())
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			tf := rt.Field(i)
			tag := tf.Tag.Get("yaml")
			if tag == "-" {
				continue
			}
			fieldTag := tf.Tag.Get("comment")
			hfieldTag := tf.Tag.Get("hcomment")
			ffieldTag := tf.Tag.Get("fcomment")
			if fieldTag != "" {
				fieldTag = fmt.Sprintf("%s <%s>", fieldTag, tf.Type.String())
			}
			if hfieldTag != "" {
				hfieldTag = fmt.Sprintf("%s <%s>", hfieldTag, tf.Type.String())
			}
			if ffieldTag != "" {
				ffieldTag = fmt.Sprintf("%s <%s>", ffieldTag, tf.Type.String())
			}
			node.Content[i*2].LineComment = fieldTag
			node.Content[i*2+1].LineComment = fieldTag
			node.Content[i*2].HeadComment = hfieldTag
			//node.Content[i*2+1].HeadComment = hfieldTag
			node.Content[i*2].FootComment = ffieldTag
			//node.Content[i*2+1].FootComment = ffieldTag
			yamlWithCommentAndType(node.Content[i*2+1], rv.Field(i), tf.Type)
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			index := rv.Index(i)
			yamlWithCommentAndType(node.Content[i], index, index.Type())
		}
	case reflect.Map:
		for i, key := range rv.MapKeys() {
			value := rv.MapIndex(key)
			valueType := value.Type()
			yamlWithCommentAndType(node.Content[i*2+1], value, valueType)
		}
	}
}

func yamlWithComment2(nodes []*yaml.Node, rv reflect.Value, rt reflect.Type) {
	switch rv.Kind() {
	case reflect.Pointer:
		yamlWithComment2(nodes, rv.Elem(), rt.Elem())
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			tf := rt.Field(i)
			fieldTag := tf.Tag.Get("comment")
			hfieldTag := tf.Tag.Get("hcomment")
			ffieldTag := tf.Tag.Get("fcomment")
			nodes[i*2].LineComment = fieldTag
			nodes[i*2+1].LineComment = fieldTag
			nodes[i*2].HeadComment = hfieldTag
			//node.Content[i*2+1].HeadComment = hfieldTag
			nodes[i*2].FootComment = ffieldTag
			//node.Content[i*2+1].FootComment = ffieldTag
			yamlWithComment2(nodes[i*2+1].Content, rv.Field(i), tf.Type)
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			index := rv.Index(i)
			yamlWithComment(nodes[i], index, index.Type())
		}
	case reflect.Map:
		for i, key := range rv.MapKeys() {
			value := rv.MapIndex(key)
			valueType := value.Type()
			yamlWithComment(nodes[i*2+1], value, valueType)
		}
	}
}
