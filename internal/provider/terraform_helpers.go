package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type pointOfContactModel struct {
	Identity types.String `tfsdk:"identity"`
	Title    types.String `tfsdk:"title"`
}

func mergeString(plan, state types.String) types.String {
	if plan.IsNull() || plan.IsUnknown() {
		return state
	}
	return plan
}

func mergeBool(plan, state types.Bool) types.Bool {
	if plan.IsNull() || plan.IsUnknown() {
		return state
	}
	return plan
}

func mergeList(plan, state types.List) types.List {
	if plan.IsNull() || plan.IsUnknown() {
		return state
	}
	return plan
}

func mergeMap(plan, state types.Map) types.Map {
	if plan.IsNull() || plan.IsUnknown() {
		return state
	}
	return plan
}

func pointOfContactObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"identity": types.StringType,
		"title":    types.StringType,
	}
}

func pointOfContactListToBodies(ctx context.Context, value types.List) []pointOfContactBody {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	var models []pointOfContactModel
	if diags := value.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil
	}
	bodies := make([]pointOfContactBody, 0, len(models))
	for _, model := range models {
		body := pointOfContactBody{}
		if !model.Identity.IsNull() && !model.Identity.IsUnknown() && strings.TrimSpace(model.Identity.ValueString()) != "" {
			identity := model.Identity.ValueString()
			body.Identity = &identity
		}
		if !model.Title.IsNull() && !model.Title.IsUnknown() && strings.TrimSpace(model.Title.ValueString()) != "" {
			title := model.Title.ValueString()
			body.Title = &title
		}
		bodies = append(bodies, body)
	}
	return bodies
}

func pointOfContactBodiesToList(values []pointOfContactBody) types.List {
	if len(values) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: pointOfContactObjectType()})
	}
	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		attributes := map[string]attr.Value{
			"identity": nullableString(value.Identity),
			"title":    nullableString(value.Title),
		}
		obj, diags := types.ObjectValue(pointOfContactObjectType(), attributes)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: pointOfContactObjectType()})
		}
		elements = append(elements, obj)
	}
	return types.ListValueMust(types.ObjectType{AttrTypes: pointOfContactObjectType()}, elements)
}

func tagsMapToBodies(values map[string]string) []tagBody {
	if len(values) == 0 {
		return nil
	}
	bodies := make([]tagBody, 0, len(values))
	for name, value := range values {
		bodies = append(bodies, tagBody{Name: name, Value: value})
	}
	return bodies
}

func tagsBodiesToMap(values []tagBody) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(values))
	for _, value := range values {
		name := strings.TrimSpace(value.Name)
		if name == "" {
			continue
		}
		out[name] = value.Value
	}
	return out
}
