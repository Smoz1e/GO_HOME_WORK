package productv1

import (
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetProductRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Sku           int64                  `protobuf:"varint,1,opt,name=sku,proto3" json:"sku,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetProductRequest) Reset() {
	*x = GetProductRequest{}
	mi := &file_product_v1_product_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetProductRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetProductRequest) ProtoMessage() {}

func (x *GetProductRequest) ProtoReflect() protoreflect.Message {
	mi := &file_product_v1_product_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetProductRequest) Descriptor() ([]byte, []int) {
	return file_product_v1_product_proto_rawDescGZIP(), []int{0}
}

func (x *GetProductRequest) GetSku() int64 {
	if x != nil {
		return x.Sku
	}
	return 0
}

type GetProductResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Sku           int64                  `protobuf:"varint,1,opt,name=sku,proto3" json:"sku,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Price         int32                  `protobuf:"varint,3,opt,name=price,proto3" json:"price,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetProductResponse) Reset() {
	*x = GetProductResponse{}
	mi := &file_product_v1_product_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetProductResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetProductResponse) ProtoMessage() {}

func (x *GetProductResponse) ProtoReflect() protoreflect.Message {
	mi := &file_product_v1_product_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetProductResponse) Descriptor() ([]byte, []int) {
	return file_product_v1_product_proto_rawDescGZIP(), []int{1}
}

func (x *GetProductResponse) GetSku() int64 {
	if x != nil {
		return x.Sku
	}
	return 0
}

func (x *GetProductResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetProductResponse) GetPrice() int32 {
	if x != nil {
		return x.Price
	}
	return 0
}

type ListProductsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Page          int32                  `protobuf:"varint,1,opt,name=page,proto3" json:"page,omitempty"`
	Limit         int32                  `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListProductsRequest) Reset() {
	*x = ListProductsRequest{}
	mi := &file_product_v1_product_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListProductsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListProductsRequest) ProtoMessage() {}

func (x *ListProductsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_product_v1_product_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ListProductsRequest) Descriptor() ([]byte, []int) {
	return file_product_v1_product_proto_rawDescGZIP(), []int{2}
}

func (x *ListProductsRequest) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *ListProductsRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

type Product struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Sku           int64                  `protobuf:"varint,1,opt,name=sku,proto3" json:"sku,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Price         int32                  `protobuf:"varint,3,opt,name=price,proto3" json:"price,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Product) Reset() {
	*x = Product{}
	mi := &file_product_v1_product_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Product) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Product) ProtoMessage() {}

func (x *Product) ProtoReflect() protoreflect.Message {
	mi := &file_product_v1_product_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Product) Descriptor() ([]byte, []int) {
	return file_product_v1_product_proto_rawDescGZIP(), []int{3}
}

func (x *Product) GetSku() int64 {
	if x != nil {
		return x.Sku
	}
	return 0
}

func (x *Product) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Product) GetPrice() int32 {
	if x != nil {
		return x.Price
	}
	return 0
}

type ListProductsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Products      []*Product             `protobuf:"bytes,1,rep,name=products,proto3" json:"products,omitempty"`
	Total         int32                  `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
	TotalPages    int32                  `protobuf:"varint,3,opt,name=total_pages,json=totalPages,proto3" json:"total_pages,omitempty"`
	Page          int32                  `protobuf:"varint,4,opt,name=page,proto3" json:"page,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListProductsResponse) Reset() {
	*x = ListProductsResponse{}
	mi := &file_product_v1_product_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListProductsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListProductsResponse) ProtoMessage() {}

func (x *ListProductsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_product_v1_product_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ListProductsResponse) Descriptor() ([]byte, []int) {
	return file_product_v1_product_proto_rawDescGZIP(), []int{4}
}

func (x *ListProductsResponse) GetProducts() []*Product {
	if x != nil {
		return x.Products
	}
	return nil
}

func (x *ListProductsResponse) GetTotal() int32 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *ListProductsResponse) GetTotalPages() int32 {
	if x != nil {
		return x.TotalPages
	}
	return 0
}

func (x *ListProductsResponse) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

var File_product_v1_product_proto protoreflect.FileDescriptor

const file_product_v1_product_proto_rawDesc = "" +
	"\n" +
	"\x18product/v1/product.proto\x12\n" +
	"product.v1\"%\n" +
	"\x11GetProductRequest\x12\x10\n" +
	"\x03sku\x18\x01 \x01(\x03R\x03sku\"P\n" +
	"\x12GetProductResponse\x12\x10\n" +
	"\x03sku\x18\x01 \x01(\x03R\x03sku\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\x12\x14\n" +
	"\x05price\x18\x03 \x01(\x05R\x05price\"?\n" +
	"\x13ListProductsRequest\x12\x12\n" +
	"\x04page\x18\x01 \x01(\x05R\x04page\x12\x14\n" +
	"\x05limit\x18\x02 \x01(\x05R\x05limit\"E\n" +
	"\aProduct\x12\x10\n" +
	"\x03sku\x18\x01 \x01(\x03R\x03sku\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\x12\x14\n" +
	"\x05price\x18\x03 \x01(\x05R\x05price\"\x92\x01\n" +
	"\x14ListProductsResponse\x12/\n" +
	"\bproducts\x18\x01 \x03(\v2\x13.product.v1.ProductR\bproducts\x12\x14\n" +
	"\x05total\x18\x02 \x01(\x05R\x05total\x12\x1f\n" +
	"\vtotal_pages\x18\x03 \x01(\x05R\n" +
	"totalPages\x12\x12\n" +
	"\x04page\x18\x04 \x01(\x05R\x04page2\xb0\x01\n" +
	"\x0eProductService\x12K\n" +
	"\n" +
	"GetProduct\x12\x1d.product.v1.GetProductRequest\x1a\x1e.product.v1.GetProductResponse\x12Q\n" +
	"\fListProducts\x12\x1f.product.v1.ListProductsRequest\x1a .product.v1.ListProductsResponseB:Z8gitlab.ozon.dev/go/classrom/HW2/gen/product/v1;productv1b\x06proto3"

var (
	file_product_v1_product_proto_rawDescOnce sync.Once
	file_product_v1_product_proto_rawDescData []byte
)

func file_product_v1_product_proto_rawDescGZIP() []byte {
	file_product_v1_product_proto_rawDescOnce.Do(func() {
		file_product_v1_product_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_product_v1_product_proto_rawDesc), len(file_product_v1_product_proto_rawDesc)))
	})
	return file_product_v1_product_proto_rawDescData
}

var file_product_v1_product_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_product_v1_product_proto_goTypes = []any{
	(*GetProductRequest)(nil),
	(*GetProductResponse)(nil),
	(*ListProductsRequest)(nil),
	(*Product)(nil),
	(*ListProductsResponse)(nil),
}
var file_product_v1_product_proto_depIdxs = []int32{
	3,
	0,
	2,
	1,
	4,
	3,
	1,
	1,
	1,
	0,
}

func init() { file_product_v1_product_proto_init() }
func file_product_v1_product_proto_init() {
	if File_product_v1_product_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_product_v1_product_proto_rawDesc), len(file_product_v1_product_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_product_v1_product_proto_goTypes,
		DependencyIndexes: file_product_v1_product_proto_depIdxs,
		MessageInfos:      file_product_v1_product_proto_msgTypes,
	}.Build()
	File_product_v1_product_proto = out.File
	file_product_v1_product_proto_goTypes = nil
	file_product_v1_product_proto_depIdxs = nil
}
