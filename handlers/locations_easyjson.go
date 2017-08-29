// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package handlers

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonE9bf8de2DecodeGithubComAgneumTravelsHandlers(in *jlexer.Lexer, out *Location) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.Id = uint32(in.Uint32())
		case "place":
			out.Place = string(in.String())
		case "country":
			out.Country = string(in.String())
		case "city":
			out.City = string(in.String())
		case "distance":
			out.Distance = uint32(in.Uint32())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonE9bf8de2EncodeGithubComAgneumTravelsHandlers(out *jwriter.Writer, in Location) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"id\":")
	out.Uint32(uint32(in.Id))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"place\":")
	out.String(string(in.Place))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"country\":")
	out.String(string(in.Country))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"city\":")
	out.String(string(in.City))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"distance\":")
	out.Uint32(uint32(in.Distance))
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Location) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonE9bf8de2EncodeGithubComAgneumTravelsHandlers(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Location) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonE9bf8de2EncodeGithubComAgneumTravelsHandlers(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Location) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonE9bf8de2DecodeGithubComAgneumTravelsHandlers(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Location) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonE9bf8de2DecodeGithubComAgneumTravelsHandlers(l, v)
}
