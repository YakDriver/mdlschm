package mdlschm

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/numbervalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model any
		want  tfsdk.Schema
	}{
		"Basic": {
			model: struct {
				Name types.String `tfsdk:"name" required:"true"`
			}{},
			want: tfsdk.Schema{
				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
		},
		"Simple": {
			model: struct {
				Name                      types.String `tfsdk:"name" required:"true"`
				DisableExecuteAPIEndpoint types.Bool   `tfsdk:"disable_execute_api_endpoint" optional:"true" computed:"true"`
				MinimumCompressionSize    int          `tfsdk:"minimum_compression_size" computed:"true"`
				PercentTraffic            float64      `tfsdk:"percent_traffic" optional:"true"`
			}{},
			want: tfsdk.Schema{
				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
					},
					"disable_execute_api_endpoint": {
						Type:     types.BoolType,
						Optional: true,
						Computed: true,
					},
					"minimum_compression_size": {
						Type:     types.Int64Type,
						Computed: true,
					},
					"percent_traffic": {
						Type:     types.Float64Type,
						Optional: true,
					},
				},
			},
		},
		"Texting": {
			model: struct {
				_                         struct{}     `md:"Description, markdown, deprecation, sensitive tests" version:1 desc:"This is your description speaking" deprecation:"Prepare for deprecation"`
				Name                      types.String `tfsdk:"name" required:"true" md:"Markdown's description"`
				DisableExecuteAPIEndpoint types.Bool   `tfsdk:"disable_execute_api_endpoint" desc:"Just a regular, old description with spaces and a comma" optional:"true" computed:"true"`
				MinimumCompressionSize    int          `tfsdk:"minimum_compression_size" computed:"true" deprecation:"This is going away"`
				PercentTraffic            float64      `tfsdk:"percent_traffic" optional:"true" sensitive:"true" md:"Markdown's description" desc:"Just a regular, old description with spaces and a comma" deprecation:"This is going away"`
			}{},
			want: tfsdk.Schema{
				MarkdownDescription: "Description, markdown, deprecation, sensitive tests",
				Description:         "This is your description speaking",
				Version:             1,
				DeprecationMessage:  "Prepare for deprecation",
				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:                types.StringType,
						Required:            true,
						MarkdownDescription: "Markdown's description",
					},
					"disable_execute_api_endpoint": {
						Type:        types.BoolType,
						Optional:    true,
						Computed:    true,
						Description: "Just a regular, old description with spaces and a comma",
					},
					"minimum_compression_size": {
						Type:               types.Int64Type,
						Computed:           true,
						DeprecationMessage: "This is going away",
					},
					"percent_traffic": {
						Type:                types.Float64Type,
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "Markdown's description",
						Description:         "Just a regular, old description with spaces and a comma",
						DeprecationMessage:  "This is going away",
					},
				},
			},
		},
		"PlanModifiers": {
			model: struct {
				Name   types.String  `tfsdk:"name" required:"true" pmods:"replace"`
				Fame   types.String  `tfsdk:"fame" required:"true" pmods:"default(game)"`
				Same   types.String  `tfsdk:"same" required:"true" pmods:"replace,default(game)"`
				Tame   types.String  `tfsdk:"tame" required:"true" pmods:"usfu"`
				Bool   types.Bool    `tfsdk:"bool" required:"true" pmods:"default(false)"`
				Float  types.Float64 `tfsdk:"float" required:"true" pmods:"default(0.0)"`
				Int    types.Int64   `tfsdk:"int" required:"true" pmods:"default(12)"`
				Number types.Number  `tfsdk:"number" required:"true" pmods:"default(1)"`
			}{},
			want: tfsdk.Schema{
				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
						},
					},
					"fame": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultValue(types.String{Value: "game"}),
						},
					},
					"same": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
							DefaultValue(types.String{Value: "game"}),
						},
					},
					"tame": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.UseStateForUnknown(),
						},
					},
					"bool": {
						Type:     types.BoolType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultValue(types.Bool{Value: false}),
						},
					},
					"float": {
						Type:     types.Float64Type,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultValue(types.Float64{Value: 0.0}),
						},
					},
					"int": {
						Type:     types.Int64Type,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultValue(types.Int64{Value: 12}),
						},
					},
					"number": {
						Type:     types.NumberType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultValue(types.Number{Value: big.NewFloat(1)}),
						},
					},
				},
			},
		},
		"Validators": {
			model: struct {
				Name types.String `tfsdk:"name" required:"true" valid:"between(3,32)"`
				ID   types.Number `tfsdk:"id" optional:"true" valid:"between(0,90)"`
			}{},
			want: tfsdk.Schema{
				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.LengthBetween(3, 32),
						},
					},
					"id": {
						Type:     types.NumberType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							float64validator.Between(0, 90),
						},
					},
				},
			},
		},
		"PlanModifiersAndValidators": {
			model: struct {
				Name types.String `tfsdk:"name" required:"true" pmods:"usfu" valid:"between(3,32)"`
				Same types.String `tfsdk:"same" optional:"true" valid:"between(0,255)" pmods:"replace,default(game)"`
				ID   types.Number `tfsdk:"id" optional:"true" valid:"between(0,90)" pmods:"usfu,replace"`
			}{},
			want: tfsdk.Schema{
				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.LengthBetween(3, 32),
						},
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.UseStateForUnknown(),
						},
					},
					"same": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.LengthBetween(0, 255),
						},
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
							DefaultValue(types.String{Value: "game"}),
						},
					},
					"id": {
						Type:     types.NumberType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							float64validator.Between(0, 90),
						},
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
							resource.UseStateForUnknown(),
						},
					},
				},
			},
		},
		"Variety": {
			model: struct {
				_                         struct{}           `md:"Test of a variety of arguments"`
				Name                      types.String       `tfsdk:"name" required:"true" pmods:"replace"`
				DisableExecuteAPIEndpoint types.Bool         `tfsdk:"disable_execute_api_endpoint" optional:"true" computed:"true"`
				MinimumCompressionSize    int64              `tfsdk:"minimum_compression_size" computed:"true"`
				PercentTraffic            float64            `tfsdk:"percent_traffic" optional:"true" pmods:"default(0)"`
				Parameters                map[string]string  `tfsdk:"parameters" optional:"true"`
				AdditionalVersionWeights  map[string]float64 `tfsdk:"additional_version_weights" optional:"true"`
				VPCEndpointIDs            []string           `tfsdk:"vpc_endpoint_ids" computed:"true" collection:"set" valid:"between(0,10)"`
				BinaryMediaTypes          []string           `tfsdk:"binary_media_types" optional:"true" valid:"between(0,10)"`
				Ports                     []types.Number     `tfsdk:"ports" optional:"true"`

				EndpointConfiguration struct {
					Field types.String `tfsdk:"field" computed:"true"`
					Other types.String `tfsdk:"other" computed:"true"`
				} `tfsdk:"endpoint_configuration"`

				Criterion []struct {
					Field types.String `tfsdk:"field" required:"true"`
					Other types.String `tfsdk:"other" optional:"true"`
				} `tfsdk:"criterion" valid:"between(0,10)" collection:"set"`
			}{},
			want: tfsdk.Schema{
				MarkdownDescription: "Test of a variety of arguments",

				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							resource.RequiresReplace(),
						},
					},
					"disable_execute_api_endpoint": {
						Type:     types.BoolType,
						Optional: true,
						Computed: true,
					},
					"minimum_compression_size": {
						Type:     types.Int64Type,
						Computed: true,
					},
					"percent_traffic": {
						Type:     types.Float64Type,
						Optional: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultValue(types.Float64{Value: 0.0}),
						},
					},
					"parameters": {
						Optional: true,
						Type: types.MapType{
							ElemType: types.StringType,
						},
					},
					"additional_version_weights": {
						Optional: true,
						Type: types.MapType{
							ElemType: types.Float64Type,
						},
					},
					"vpc_endpoint_ids": {
						Type: types.SetType{
							ElemType: types.StringType,
						},
						Computed: true,
						Validators: []tfsdk.AttributeValidator{
							setvalidator.SizeBetween(0, 10),
						},
					},
					"binary_media_types": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							listvalidator.SizeBetween(0, 10),
						},
					},
					"ports": {
						Type: types.ListType{
							ElemType: types.NumberType,
						},
						Optional: true,
					},
				},
				Blocks: map[string]tfsdk.Block{
					"endpoint_configuration": {
						NestingMode: tfsdk.BlockNestingModeList,
						Validators: []tfsdk.AttributeValidator{
							listvalidator.SizeBetween(0, 1),
						},
						Attributes: map[string]tfsdk.Attribute{
							"field": {
								Computed: true,
								Type:     types.StringType,
							},
							"other": {
								Computed: true,
								Type:     types.StringType,
							},
						},
					},
					"criterion": {
						NestingMode: tfsdk.BlockNestingModeSet,
						Validators: []tfsdk.AttributeValidator{
							setvalidator.SizeBetween(0, 10),
						},
						Attributes: map[string]tfsdk.Attribute{
							"field": {
								Required: true,
								Type:     types.StringType,
							},
							"other": {
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
		},
		"BasicBlock": {
			model: struct {
				Endpoint struct {
					Field types.String `tfsdk:"field" required:"true"`
					Other types.String `tfsdk:"other" optional:"true"`
				} `tfsdk:"endpoint"`
			}{},
			want: tfsdk.Schema{
				Blocks: map[string]tfsdk.Block{
					"endpoint": {
						NestingMode: tfsdk.BlockNestingModeList,
						Validators: []tfsdk.AttributeValidator{
							listvalidator.SizeBetween(0, 1),
						},
						Attributes: map[string]tfsdk.Attribute{
							"field": {
								Required: true,
								Type:     types.StringType,
							},
							"other": {
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
		},
		"BasicBlocks": {
			model: struct {
				Endpoint struct {
					Field types.String `tfsdk:"field" required:"true"`
					Other types.String `tfsdk:"other" optional:"true"`
				} `tfsdk:"endpoint"`
				Bright struct {
					Zeds  types.String `tfsdk:"zeds" optional:"true"`
					Dead  types.String `tfsdk:"dead" optional:"true"`
					Alive types.String `tfsdk:"alive" optional:"true"`
				} `tfsdk:"bright"`
			}{},
			want: tfsdk.Schema{
				Blocks: map[string]tfsdk.Block{
					"endpoint": {
						NestingMode: tfsdk.BlockNestingModeList,
						Validators: []tfsdk.AttributeValidator{
							listvalidator.SizeBetween(0, 1),
						},
						Attributes: map[string]tfsdk.Attribute{
							"field": {
								Required: true,
								Type:     types.StringType,
							},
							"other": {
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
					"bright": {
						NestingMode: tfsdk.BlockNestingModeList,
						Validators: []tfsdk.AttributeValidator{
							listvalidator.SizeBetween(0, 1),
						},
						Attributes: map[string]tfsdk.Attribute{
							"zeds": {
								Optional: true,
								Type:     types.StringType,
							},
							"dead": {
								Optional: true,
								Type:     types.StringType,
							},
							"alive": {
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
		},
		"BasicBlockList": {
			model: struct {
				Endpoint []struct {
					Field types.String `tfsdk:"field" required:"true"`
					Other types.String `tfsdk:"other" optional:"true"`
				} `tfsdk:"endpoint"`
			}{},
			want: tfsdk.Schema{
				Blocks: map[string]tfsdk.Block{
					"endpoint": {
						NestingMode: tfsdk.BlockNestingModeList,
						Attributes: map[string]tfsdk.Attribute{
							"field": {
								Required: true,
								Type:     types.StringType,
							},
							"other": {
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
		},
		"BasicBlockSet": {
			model: struct {
				Endpoint []struct {
					Field types.String `tfsdk:"field" required:"true"`
					Other types.String `tfsdk:"other" optional:"true"`
				} `tfsdk:"endpoint" collection:"set"`
			}{},
			want: tfsdk.Schema{
				Blocks: map[string]tfsdk.Block{
					"endpoint": {
						NestingMode: tfsdk.BlockNestingModeSet,
						Attributes: map[string]tfsdk.Attribute{
							"field": {
								Required: true,
								Type:     types.StringType,
							},
							"other": {
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
				},
			},
		},
		"Nesting": {
			model: struct {
				Endpoint struct {
					Field struct {
						InnerField []struct {
							FarInField types.String `tfsdk:"far_in_field" required:"true"`
							FarInOther types.String `tfsdk:"far_in_other" optional:"true"`
						} `tfsdk:"inner_field" collection:"set"`
						InnerOther types.String `tfsdk:"inner_other" optional:"true"`
					} `tfsdk:"field" required:"true"`
					Other types.String `tfsdk:"other" optional:"true"`
				} `tfsdk:"endpoint"`
			}{},
			want: tfsdk.Schema{
				Blocks: map[string]tfsdk.Block{
					"endpoint": {
						NestingMode: tfsdk.BlockNestingModeList,
						Validators: []tfsdk.AttributeValidator{
							listvalidator.SizeBetween(0, 1),
						},
						Attributes: map[string]tfsdk.Attribute{
							"other": {
								Optional: true,
								Type:     types.StringType,
							},
						},
						Blocks: map[string]tfsdk.Block{
							"field": {
								NestingMode: tfsdk.BlockNestingModeList,
								Validators: []tfsdk.AttributeValidator{
									listvalidator.SizeBetween(1, 1),
								},
								Attributes: map[string]tfsdk.Attribute{
									"inner_other": {
										Optional: true,
										Type:     types.StringType,
									},
								},
								Blocks: map[string]tfsdk.Block{
									"inner_field": {
										NestingMode: tfsdk.BlockNestingModeSet,
										Attributes: map[string]tfsdk.Attribute{
											"far_in_field": {
												Required: true,
												Type:     types.StringType,
											},
											"far_in_other": {
												Optional: true,
												Type:     types.StringType,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"Between": {
			model: struct {
				Bynx     []types.String `required:"true" valid:"between(1,10)"`
				Dahlback []types.String `collection:"set" valid:"between(0,3)"`
				Shallou  types.String   `optional:"true" valid:"between(3,8)"`
				Dekleyn  types.Float64  `valid:"between(4,5)"`
				AMR      int            `valid:"between(1,100)"`
			}{},
			want: tfsdk.Schema{
				Attributes: map[string]tfsdk.Attribute{
					"bynx": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							listvalidator.SizeBetween(1, 10),
						},
					},
					"dahlback": {
						Type: types.SetType{
							ElemType: types.StringType,
						},
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							setvalidator.SizeBetween(0, 3),
						},
					},
					"shallou": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.LengthBetween(3, 8),
						},
					},
					"dekleyn": {
						Type:     types.Float64Type,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							float64validator.Between(4, 5),
						},
					},
					"amr": {
						Type:     types.Int64Type,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							int64validator.Between(1, 100),
						},
					},
				},
			},
		},
		"OneOf": {
			model: struct {
				Name types.String  `required:"true" valid:"oneof(sultan,shepard,ben,böhmer)"`
				Bits types.Number  `optional:"true" valid:"oneof(0,8,24,64)"`
				Size types.Float64 `valid:"oneof(2.1,84.5,240.1,649.123)"`
				Mode int           `valid:"oneof(1,2,5,13)"`
			}{},
			want: tfsdk.Schema{
				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf("sultan", "shepard", "ben", "böhmer"),
						},
					},
					"bits": {
						Type:     types.NumberType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							numbervalidator.OneOf([]*big.Float{
								big.NewFloat(0),
								big.NewFloat(8),
								big.NewFloat(24),
								big.NewFloat(64),
							}...),
						},
					},
					"size": {
						Type:     types.Float64Type,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							float64validator.OneOf(2.1, 84.5, 240.1, 649.123),
						},
					},
					"mode": {
						Type:     types.Int64Type,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							int64validator.OneOf(1, 2, 5, 13),
						},
					},
				},
			},
		},
		"NoneOf": {
			model: struct {
				Name types.String  `required:"true" valid:"noneof(sultan,shepard,ben,böhmer)"`
				Bits types.Number  `optional:"true" valid:"noneof(0,8,24,64)"`
				Size types.Float64 `valid:"noneof(2.1,84.5,240.1,649.123)"`
				Mode int           `valid:"noneof(1,2,5,13)"`
			}{},
			want: tfsdk.Schema{
				Attributes: map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.NoneOf("sultan", "shepard", "ben", "böhmer"),
						},
					},
					"bits": {
						Type:     types.NumberType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							numbervalidator.NoneOf([]*big.Float{
								big.NewFloat(0),
								big.NewFloat(8),
								big.NewFloat(24),
								big.NewFloat(64),
							}...),
						},
					},
					"size": {
						Type:     types.Float64Type,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							float64validator.NoneOf(2.1, 84.5, 240.1, 649.123),
						},
					},
					"mode": {
						Type:     types.Int64Type,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							int64validator.NoneOf(1, 2, 5, 13),
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := New(test.model)

			diff := deep.Equal(got, test.want)
			if diff != nil {
				t.Errorf("got: %+v\nwant: %+v\ndifference: %v", got, test.want, diff)
			}
		})
	}
}

func TestSplitTagValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args string
		want []string
	}{
		"Basic": {
			args: `between(3,32)`,
			want: []string{
				"between(3,32)",
			},
		},
		"Test2": {
			args: `between(3,32),arbitrary(5,2,3,1)`,
			want: []string{
				"between(3,32)",
				"arbitrary(5,2,3,1)",
			},
		},
		"Test3": {
			args: `fred,any(5),toast(1,8),between(3,32),arbitrary(5,2,3,1)`,
			want: []string{
				"fred",
				"any(5)",
				"toast(1,8)",
				"between(3,32)",
				"arbitrary(5,2,3,1)",
			},
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := splitTagValues(test.args)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected difference:\ngot %+v\nexpected %+v", got, test.want)
			}
		})
	}
}

func TestSplitTags(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		tags string
		want []string
	}{
		"Basic": {
			tags: `tfsdk:"name" required:"true" valid:"between(3,32)"`,
			want: []string{
				`tfsdk:"name"`,
				`required:"true"`,
				`valid:"between(3,32)"`,
			},
		},
		"Test2": {
			tags: `tfsdk:"name" required:"true" valid:"between(3, 32)"`,
			want: []string{
				`tfsdk:"name"`,
				`required:"true"`,
				`valid:"between(3, 32)"`,
			},
		},
		"Test3": {
			tags: `tfsdk:"name" md:"This is a description with spaces" valid:"between( 3, 32 )"`,
			want: []string{
				`tfsdk:"name"`,
				`md:"This is a description with spaces"`,
				`valid:"between( 3, 32 )"`,
			},
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := splitTags(test.tags)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected difference:\ngot %+v\nexpected %+v", got, test.want)
			}
		})
	}
}

func TestSnakeCase(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		camel string
		want  string
	}{
		"Basic": {
			camel: "SuperSimpleCase",
			want:  "super_simple_case",
		},
		"Caps": {
			camel: "VPCName",
			want:  "vpc_name",
		},
		"Caps2": {
			camel: "LittleVPCName",
			want:  "little_vpc_name",
		},
		"Caps3": {
			camel: "VPCEndpointIDs",
			want:  "vpc_endpoint_ids",
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := snakeCase(test.camel, "")

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("unexpected difference:\ngot %+v\nexpected %+v", got, test.want)
			}
		})
	}
}
