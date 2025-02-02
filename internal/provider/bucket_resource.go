// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &orderResource{}
	_ resource.ResourceWithConfigure = &orderResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewOrderResource() resource.Resource {
	return &orderResource{}
}

// orderResource is the resource implementation.
type orderResource struct {
	client *session.Session
}

// orderResourceModel maps the resource schema data.
type orderResourceModel struct {
	Date        tftypes.String `tfsdk:"date"`
	Name        tftypes.String `tfsdk:"name"`
	Tags        tftypes.String `tfsdk:"tags"`
	LastUpdated tftypes.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *orderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

// Schema defines the schema for the resource.
func (r *orderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an s3 bucket.",
		Attributes: map[string]schema.Attribute{
			"last_updated": schema.StringAttribute{
				Description: "Date Updated",
				Computed:    true,
			},
			"date": schema.StringAttribute{
				Description: "Date Created",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "S3 Bucket Name",
				Required:    true,
			},
			"tags": schema.StringAttribute{
				Description: "S3 Bucket Tags",
				Required:    true,
			},
		},
	}
}

// Create a new resource.
func (r *orderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan orderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// for index, item := range plan.Buckets {
	// Create an S3 service client
	svc := s3.New(r.client)
	awsStringBucket := strings.Replace(plan.Name.String(), "\"", "", -1)

	// Create input parameters for the CreateBucket operation
	input := &s3.CreateBucketInput{
		Bucket: aws.String(awsStringBucket),
	}

	// Execute the CreateBucket operation
	_, err := svc.CreateBucket(input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	// Add tags
	var tags []*s3.Tag
	tagValue := strings.Replace(plan.Tags.String(), "\"", "", -1)
	tags = append(tags, &s3.Tag{
		Key:   aws.String("tfkey"),
		Value: aws.String(tagValue),
	})

	_, err = svc.PutBucketTagging(&s3.PutBucketTaggingInput{
		Bucket: aws.String(awsStringBucket),
		Tagging: &s3.Tagging{
			TagSet: tags,
		},
	})
	if err != nil {
		fmt.Println("Error adding tags to the bucket:", err)
		return
	}

	fmt.Printf("Bucket %s created successfully\n", plan.Name)

	plan.Name = types.StringValue(awsStringBucket)
	plan.Date = types.StringValue(time.Now().Format(time.RFC850))
	plan.Tags = types.StringValue(tagValue)

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *orderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state orderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsStringBucket := strings.Replace(state.Name.String(), "\"", "", -1)

	svc := s3.New(r.client)
	params := &s3.HeadBucketInput{
		Bucket: aws.String(awsStringBucket),
	}

	_, err := svc.HeadBucket(params)
	if err != nil {
		fmt.Println("Error getting bucket information:", err)
		os.Exit(1)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *orderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan orderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create an S3 service client
	svc := s3.New(r.client)
	awsStringBucket := strings.Replace(plan.Name.String(), "\"", "", -1)

	// Add tags
	var tags []*s3.Tag
	tagValue := strings.Replace(plan.Tags.String(), "\"", "", -1)
	tags = append(tags, &s3.Tag{
		Key:   aws.String("tfkey"),
		Value: aws.String(tagValue),
	})

	_, err := svc.PutBucketTagging(&s3.PutBucketTaggingInput{
		Bucket: aws.String(awsStringBucket),
		Tagging: &s3.Tagging{
			TagSet: tags,
		},
	})
	if err != nil {
		fmt.Println("Error adding tags to the bucket:", err)
		return
	}

	plan.Name = types.StringValue(strings.Replace(awsStringBucket, "\"", "", -1))
	plan.Date = types.StringValue(time.Now().Format(time.RFC850))
	plan.Tags = types.StringValue(strings.Replace(tagValue, "\"", "", -1))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *orderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state orderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc := s3.New(r.client)

	input := &s3.DeleteBucketInput{
		Bucket: aws.String(strings.Replace(state.Name.String(), "\"", "", -1)),
	}

	_, err := svc.DeleteBucket(input)
	if err != nil {
		log.Fatalf("failed to delete bucket, %v", err)
	}
}

// Configure adds the provider configured client to the resource.
func (r *orderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*session.Session)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *session.Session, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}
