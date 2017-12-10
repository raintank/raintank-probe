package msg

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import "github.com/tinylib/msgp/msgp"

// DecodeMsg implements msgp.Decodable
func (z *Format) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zxvk uint8
		zxvk, err = dc.ReadUint8()
		(*z) = Format(zxvk)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Format) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteUint8(uint8(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Format) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendUint8(o, uint8(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Format) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zbzg uint8
		zbzg, bts, err = msgp.ReadUint8Bytes(bts)
		(*z) = Format(zbzg)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Format) Msgsize() (s int) {
	s = msgp.Uint8Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ProbeEvent) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zajw uint32
	zajw, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zajw > 0 {
		zajw--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Id":
			z.Id, err = dc.ReadString()
			if err != nil {
				return
			}
		case "EventType":
			z.EventType, err = dc.ReadString()
			if err != nil {
				return
			}
		case "OrgId":
			z.OrgId, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "Severity":
			z.Severity, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Source":
			z.Source, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Timestamp":
			z.Timestamp, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "Message":
			z.Message, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Tags":
			var zwht uint32
			zwht, err = dc.ReadMapHeader()
			if err != nil {
				return
			}
			if z.Tags == nil && zwht > 0 {
				z.Tags = make(map[string]string, zwht)
			} else if len(z.Tags) > 0 {
				for key, _ := range z.Tags {
					delete(z.Tags, key)
				}
			}
			for zwht > 0 {
				zwht--
				var zbai string
				var zcmr string
				zbai, err = dc.ReadString()
				if err != nil {
					return
				}
				zcmr, err = dc.ReadString()
				if err != nil {
					return
				}
				z.Tags[zbai] = zcmr
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ProbeEvent) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 8
	// write "Id"
	err = en.Append(0x88, 0xa2, 0x49, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Id)
	if err != nil {
		return
	}
	// write "EventType"
	err = en.Append(0xa9, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.EventType)
	if err != nil {
		return
	}
	// write "OrgId"
	err = en.Append(0xa5, 0x4f, 0x72, 0x67, 0x49, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.OrgId)
	if err != nil {
		return
	}
	// write "Severity"
	err = en.Append(0xa8, 0x53, 0x65, 0x76, 0x65, 0x72, 0x69, 0x74, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Severity)
	if err != nil {
		return
	}
	// write "Source"
	err = en.Append(0xa6, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Source)
	if err != nil {
		return
	}
	// write "Timestamp"
	err = en.Append(0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Timestamp)
	if err != nil {
		return
	}
	// write "Message"
	err = en.Append(0xa7, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Message)
	if err != nil {
		return
	}
	// write "Tags"
	err = en.Append(0xa4, 0x54, 0x61, 0x67, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteMapHeader(uint32(len(z.Tags)))
	if err != nil {
		return
	}
	for zbai, zcmr := range z.Tags {
		err = en.WriteString(zbai)
		if err != nil {
			return
		}
		err = en.WriteString(zcmr)
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ProbeEvent) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 8
	// string "Id"
	o = append(o, 0x88, 0xa2, 0x49, 0x64)
	o = msgp.AppendString(o, z.Id)
	// string "EventType"
	o = append(o, 0xa9, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65)
	o = msgp.AppendString(o, z.EventType)
	// string "OrgId"
	o = append(o, 0xa5, 0x4f, 0x72, 0x67, 0x49, 0x64)
	o = msgp.AppendInt64(o, z.OrgId)
	// string "Severity"
	o = append(o, 0xa8, 0x53, 0x65, 0x76, 0x65, 0x72, 0x69, 0x74, 0x79)
	o = msgp.AppendString(o, z.Severity)
	// string "Source"
	o = append(o, 0xa6, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65)
	o = msgp.AppendString(o, z.Source)
	// string "Timestamp"
	o = append(o, 0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	o = msgp.AppendInt64(o, z.Timestamp)
	// string "Message"
	o = append(o, 0xa7, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	o = msgp.AppendString(o, z.Message)
	// string "Tags"
	o = append(o, 0xa4, 0x54, 0x61, 0x67, 0x73)
	o = msgp.AppendMapHeader(o, uint32(len(z.Tags)))
	for zbai, zcmr := range z.Tags {
		o = msgp.AppendString(o, zbai)
		o = msgp.AppendString(o, zcmr)
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ProbeEvent) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zhct uint32
	zhct, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zhct > 0 {
		zhct--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Id":
			z.Id, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "EventType":
			z.EventType, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "OrgId":
			z.OrgId, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "Severity":
			z.Severity, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Source":
			z.Source, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Timestamp":
			z.Timestamp, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "Message":
			z.Message, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Tags":
			var zcua uint32
			zcua, bts, err = msgp.ReadMapHeaderBytes(bts)
			if err != nil {
				return
			}
			if z.Tags == nil && zcua > 0 {
				z.Tags = make(map[string]string, zcua)
			} else if len(z.Tags) > 0 {
				for key, _ := range z.Tags {
					delete(z.Tags, key)
				}
			}
			for zcua > 0 {
				var zbai string
				var zcmr string
				zcua--
				zbai, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				zcmr, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				z.Tags[zbai] = zcmr
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ProbeEvent) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.Id) + 10 + msgp.StringPrefixSize + len(z.EventType) + 6 + msgp.Int64Size + 9 + msgp.StringPrefixSize + len(z.Severity) + 7 + msgp.StringPrefixSize + len(z.Source) + 10 + msgp.Int64Size + 8 + msgp.StringPrefixSize + len(z.Message) + 5 + msgp.MapHeaderSize
	if z.Tags != nil {
		for zbai, zcmr := range z.Tags {
			_ = zcmr
			s += msgp.StringPrefixSize + len(zbai) + msgp.StringPrefixSize + len(zcmr)
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ProbeEventJson) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zlqf uint32
	zlqf, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zlqf > 0 {
		zlqf--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Id":
			z.Id, err = dc.ReadString()
			if err != nil {
				return
			}
		case "EventType":
			z.EventType, err = dc.ReadString()
			if err != nil {
				return
			}
		case "OrgId":
			z.OrgId, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "Severity":
			z.Severity, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Source":
			z.Source, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Timestamp":
			z.Timestamp, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "Message":
			z.Message, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Tags":
			var zdaf uint32
			zdaf, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Tags) >= int(zdaf) {
				z.Tags = (z.Tags)[:zdaf]
			} else {
				z.Tags = make([]string, zdaf)
			}
			for zxhx := range z.Tags {
				z.Tags[zxhx], err = dc.ReadString()
				if err != nil {
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ProbeEventJson) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 8
	// write "Id"
	err = en.Append(0x88, 0xa2, 0x49, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Id)
	if err != nil {
		return
	}
	// write "EventType"
	err = en.Append(0xa9, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.EventType)
	if err != nil {
		return
	}
	// write "OrgId"
	err = en.Append(0xa5, 0x4f, 0x72, 0x67, 0x49, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.OrgId)
	if err != nil {
		return
	}
	// write "Severity"
	err = en.Append(0xa8, 0x53, 0x65, 0x76, 0x65, 0x72, 0x69, 0x74, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Severity)
	if err != nil {
		return
	}
	// write "Source"
	err = en.Append(0xa6, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Source)
	if err != nil {
		return
	}
	// write "Timestamp"
	err = en.Append(0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Timestamp)
	if err != nil {
		return
	}
	// write "Message"
	err = en.Append(0xa7, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Message)
	if err != nil {
		return
	}
	// write "Tags"
	err = en.Append(0xa4, 0x54, 0x61, 0x67, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.Tags)))
	if err != nil {
		return
	}
	for zxhx := range z.Tags {
		err = en.WriteString(z.Tags[zxhx])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ProbeEventJson) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 8
	// string "Id"
	o = append(o, 0x88, 0xa2, 0x49, 0x64)
	o = msgp.AppendString(o, z.Id)
	// string "EventType"
	o = append(o, 0xa9, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65)
	o = msgp.AppendString(o, z.EventType)
	// string "OrgId"
	o = append(o, 0xa5, 0x4f, 0x72, 0x67, 0x49, 0x64)
	o = msgp.AppendInt64(o, z.OrgId)
	// string "Severity"
	o = append(o, 0xa8, 0x53, 0x65, 0x76, 0x65, 0x72, 0x69, 0x74, 0x79)
	o = msgp.AppendString(o, z.Severity)
	// string "Source"
	o = append(o, 0xa6, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65)
	o = msgp.AppendString(o, z.Source)
	// string "Timestamp"
	o = append(o, 0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	o = msgp.AppendInt64(o, z.Timestamp)
	// string "Message"
	o = append(o, 0xa7, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65)
	o = msgp.AppendString(o, z.Message)
	// string "Tags"
	o = append(o, 0xa4, 0x54, 0x61, 0x67, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Tags)))
	for zxhx := range z.Tags {
		o = msgp.AppendString(o, z.Tags[zxhx])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ProbeEventJson) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zpks uint32
	zpks, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zpks > 0 {
		zpks--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Id":
			z.Id, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "EventType":
			z.EventType, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "OrgId":
			z.OrgId, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "Severity":
			z.Severity, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Source":
			z.Source, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Timestamp":
			z.Timestamp, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "Message":
			z.Message, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Tags":
			var zjfb uint32
			zjfb, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Tags) >= int(zjfb) {
				z.Tags = (z.Tags)[:zjfb]
			} else {
				z.Tags = make([]string, zjfb)
			}
			for zxhx := range z.Tags {
				z.Tags[zxhx], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ProbeEventJson) Msgsize() (s int) {
	s = 1 + 3 + msgp.StringPrefixSize + len(z.Id) + 10 + msgp.StringPrefixSize + len(z.EventType) + 6 + msgp.Int64Size + 9 + msgp.StringPrefixSize + len(z.Severity) + 7 + msgp.StringPrefixSize + len(z.Source) + 10 + msgp.Int64Size + 8 + msgp.StringPrefixSize + len(z.Message) + 5 + msgp.ArrayHeaderSize
	for zxhx := range z.Tags {
		s += msgp.StringPrefixSize + len(z.Tags[zxhx])
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ProbeEvents) DecodeMsg(dc *msgp.Reader) (err error) {
	var zrsw uint32
	zrsw, err = dc.ReadArrayHeader()
	if err != nil {
		return
	}
	if cap((*z)) >= int(zrsw) {
		(*z) = (*z)[:zrsw]
	} else {
		(*z) = make(ProbeEvents, zrsw)
	}
	for zeff := range *z {
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				return
			}
			(*z)[zeff] = nil
		} else {
			if (*z)[zeff] == nil {
				(*z)[zeff] = new(ProbeEvent)
			}
			err = (*z)[zeff].DecodeMsg(dc)
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ProbeEvents) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		return
	}
	for zxpk := range z {
		if z[zxpk] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = z[zxpk].EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ProbeEvents) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zxpk := range z {
		if z[zxpk] == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = z[zxpk].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ProbeEvents) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zobc uint32
	zobc, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		return
	}
	if cap((*z)) >= int(zobc) {
		(*z) = (*z)[:zobc]
	} else {
		(*z) = make(ProbeEvents, zobc)
	}
	for zdnj := range *z {
		if msgp.IsNil(bts) {
			bts, err = msgp.ReadNilBytes(bts)
			if err != nil {
				return
			}
			(*z)[zdnj] = nil
		} else {
			if (*z)[zdnj] == nil {
				(*z)[zdnj] = new(ProbeEvent)
			}
			bts, err = (*z)[zdnj].UnmarshalMsg(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ProbeEvents) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zsnv := range z {
		if z[zsnv] == nil {
			s += msgp.NilSize
		} else {
			s += z[zsnv].Msgsize()
		}
	}
	return
}
