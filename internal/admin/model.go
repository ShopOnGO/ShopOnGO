package main

import (
	pb "github.com/ShopOnGO/admin-proto/pkg/service"
)

type GRPCClients struct {
	CategoryClient pb.CategoryServiceClient
	BrandClient    pb.BrandServiceClient
	LinkClient     pb.LinkServiceClient
	ProductClient  pb.ProductServiceClient
	UserClient     pb.UserServiceClient
	StatClient     pb.StatServiceClient
	HomeClient     pb.HomeServiceClient
}
