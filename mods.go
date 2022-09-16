package mdlschm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// defaultValuePlanModifier specifies a default value (attr.Value) for an attribute.
type defaultValuePlanModifier struct {
	DefaultValue attr.Value
}

func DefaultValue(v attr.Value) tfsdk.AttributePlanModifier {
	return &defaultValuePlanModifier{v}
}

var _ tfsdk.AttributePlanModifier = (*defaultValuePlanModifier)(nil)

func (apm *defaultValuePlanModifier) Description(ctx context.Context) string {
	return apm.MarkdownDescription(ctx)
}

func (apm *defaultValuePlanModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Sets the default value %q (%s) if the attribute is not set", apm.DefaultValue, apm.DefaultValue.Type(ctx))
}

func (apm *defaultValuePlanModifier) Modify(_ context.Context, req tfsdk.ModifyAttributePlanRequest, res *tfsdk.ModifyAttributePlanResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.AttributeConfig.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.AttributePlan.IsUnknown() && !req.AttributePlan.IsNull() {
		return
	}

	res.AttributePlan = apm.DefaultValue
}
