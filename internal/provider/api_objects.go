package provider

import (
	"context"
	"fmt"
)

type pointOfContactBody struct {
	Identity *string `json:"identity,omitempty"`
	Title    *string `json:"title,omitempty"`
}

type engagementScopeBody struct {
	ApplicationKeys []string `json:"applicationKeys,omitempty"`
	RepositoryKeys  []string `json:"repositoryKeys,omitempty"`
}

type applicationBody struct {
	Key                    *string              `json:"key,omitempty"`
	ApiGatewayUrls         []string             `json:"apiGatewayUrls,omitempty"`
	ApiGroupKeys           []string             `json:"apiGroupKeys,omitempty"`
	ApplicationType        *string              `json:"applicationType,omitempty"`
	ApplicationTypeOther   *string              `json:"applicationTypeOther,omitempty"`
	BusinessImpact         *string              `json:"businessImpact,omitempty"`
	BusinessUnit           *string              `json:"businessUnit,omitempty"`
	ComplianceRequirements []string             `json:"complianceRequirements,omitempty"`
	DeploymentLocation     *string              `json:"deploymentLocation,omitempty"`
	Description            *string              `json:"description,omitempty"`
	EntryPoints            []string             `json:"entryPoints,omitempty"`
	EstimatedRevenue       *string              `json:"estimatedRevenue,omitempty"`
	EstimatedUsersNumber   *string              `json:"estimatedUsersNumber,omitempty"`
	IsInternetFacing       bool                 `json:"isInternetFacing"`
	Name                   *string              `json:"name,omitempty"`
	PointsOfContact        []pointOfContactBody `json:"pointsOfContact,omitempty"`
	ProjectUrls            []string             `json:"projectUrls,omitempty"`
	RepositoryUrls         []string             `json:"repositoryUrls,omitempty"`
	Tags                   []tagBody            `json:"tags,omitempty"`
	RiskScore              *int                 `json:"riskScore,omitempty"`
}

type applicationGroupBody struct {
	Key             *string              `json:"key,omitempty"`
	Applications    []string             `json:"applications,omitempty"`
	Name            *string              `json:"name,omitempty"`
	PointsOfContact []pointOfContactBody `json:"pointsOfContact,omitempty"`
	Tags            []string             `json:"tags,omitempty"`
}

type orgTeamBody struct {
	Key             *string              `json:"key,omitempty"`
	Applications    []string             `json:"applications,omitempty"`
	ApplicationTags []tagBody            `json:"applicationTags,omitempty"`
	Description     *string              `json:"description,omitempty"`
	Name            *string              `json:"name,omitempty"`
	ParentKey       *string              `json:"parentKey,omitempty"`
	PointsOfContact []pointOfContactBody `json:"pointsOfContact,omitempty"`
	ProjectUrls     []string             `json:"projectUrls,omitempty"`
	RepositoryUrls  []string             `json:"repositoryUrls,omitempty"`
	Tags            []tagBody            `json:"tags,omitempty"`
	RiskScore       *int                 `json:"riskScore,omitempty"`
}

type roleScopeBody struct {
	ApplicationKeys []string `json:"applicationKeys,omitempty"`
	OrgTeamKeys     []string `json:"orgTeamKeys,omitempty"`
	RepositoryUrls  []string `json:"repositoryUrls,omitempty"`
}

type roleBody struct {
	Key            *string           `json:"key,omitempty"`
	ApiiroGroupIds []string          `json:"apiiroGroupIds,omitempty"`
	Description    *string           `json:"description,omitempty"`
	IdpGroupIds    []string          `json:"idpGroupIds,omitempty"`
	Name           string            `json:"name,omitempty"`
	Permissions    map[string]string `json:"permissions,omitempty"`
	Scope          *roleScopeBody    `json:"scope,omitempty"`
}

type roleListItem struct {
	Key  *string `json:"key,omitempty"`
	Name *string `json:"name,omitempty"`
}

type roleGroupBody struct {
	Key         *string  `json:"key,omitempty"`
	AdminIDs    []string `json:"adminIds,omitempty"`
	Description *string  `json:"description,omitempty"`
	MemberIDs   []string `json:"memberIds,omitempty"`
	Name        string   `json:"name,omitempty"`
}

type roleGroupReference struct {
	Key         *string `json:"key,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type engagementBody struct {
	Attachments       []string             `json:"attachments,omitempty"`
	EndDate           *string              `json:"endDate,omitempty"`
	EngagementLeadKey *string              `json:"engagementLeadKey,omitempty"`
	ExternalURL       *string              `json:"externalUrl,omitempty"`
	Key               *string              `json:"key,omitempty"`
	Name              *string              `json:"name,omitempty"`
	Provider          *string              `json:"provider,omitempty"`
	Reporter          *string              `json:"reporter,omitempty"`
	Scope             *engagementScopeBody `json:"scope,omitempty"`
	StartDate         string               `json:"startDate,omitempty"`
	Status            *string              `json:"status,omitempty"`
	Summary           *string              `json:"summary,omitempty"`
	Tags              map[string]string    `json:"tags,omitempty"`
	Type              *string              `json:"type,omitempty"`
	UpdatedAt         *string              `json:"updatedAt,omitempty"`
}

type tokenPagedResponse[T any] struct {
	Next  map[string]string `json:"next"`
	Items []T               `json:"items"`
}

type tagBodyResponse struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

type applicationProfileBody struct {
	ApplicationType   *string           `json:"applicationType,omitempty"`
	BusinessImpact    *string           `json:"businessImpact,omitempty"`
	BusinessUnit      *string           `json:"businessUnit,omitempty"`
	Description       *string           `json:"description,omitempty"`
	IsActive          bool              `json:"isActive"`
	IsDeployed        bool              `json:"isDeployed"`
	IsInternetExposed bool              `json:"isInternetExposed"`
	IsPublic          bool              `json:"isPublic"`
	IsUserFacing      bool              `json:"isUserFacing"`
	Key               *string           `json:"key,omitempty"`
	Languages         []string          `json:"languages,omitempty"`
	Licenses          []string          `json:"licenses,omitempty"`
	Name              *string           `json:"name,omitempty"`
	RiskLevel         *string           `json:"riskLevel,omitempty"`
	RiskScore         *int              `json:"riskScore,omitempty"`
	Tags              []tagBodyResponse `json:"tags,omitempty"`
}

type auditLogBody struct {
	ErrorDescription   *string `json:"errorDescription,omitempty"`
	EventDescription   *string `json:"eventDescription,omitempty"`
	EventType          *string `json:"eventType,omitempty"`
	ImpactedEntityID   *string `json:"impactedEntityId,omitempty"`
	ImpactedEntityType *string `json:"impactedEntityType,omitempty"`
	ImpactedEntityURL  *string `json:"impactedEntityUrl,omitempty"`
	Key                *string `json:"key,omitempty"`
	SourceIPAddress    *string `json:"sourceIpAddress,omitempty"`
	Status             *string `json:"status,omitempty"`
	Time               *string `json:"time,omitempty"`
	Timezone           *string `json:"timezone,omitempty"`
	User               *string `json:"user,omitempty"`
	UserAgent          *string `json:"userAgent,omitempty"`
	UserType           *string `json:"userType,omitempty"`
}

type connectorBody struct {
	ID                  *string `json:"id,omitempty"`
	Provider            *string `json:"provider,omitempty"`
	TokenExpirationDate *string `json:"tokenExpirationDate,omitempty"`
	URL                 *string `json:"url,omitempty"`
}

func (c *Client) listApplications(ctx context.Context) ([]applicationBody, error) {
	const pageSize = 100
	all := make([]applicationBody, 0)
	skip := 0
	for {
		endpoint := fmt.Sprintf("/rest-api/v1/applications?skip=%d&pageSize=%d", skip, pageSize)
		var out apiPagedResponse[applicationBody]
		if err := c.doJSON(ctx, "GET", endpoint, nil, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Items...)
		skip += len(out.Items)
		if len(out.Items) == 0 || skip >= out.Paging.TotalItemCount {
			break
		}
	}
	return all, nil
}

func (c *Client) getApplication(ctx context.Context, key string) (*applicationBody, error) {
	var out applicationBody
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/rest-api/v1/applications/%s", pathEscape(key)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) createApplication(ctx context.Context, body applicationBody) (string, error) {
	var out string
	if err := c.doJSON(ctx, "POST", "/rest-api/v1/applications", body, &out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Client) updateApplication(ctx context.Context, key string, body applicationBody) error {
	return c.doJSON(ctx, "PUT", fmt.Sprintf("/rest-api/v1/applications/%s", pathEscape(key)), body, nil)
}

func (c *Client) deleteApplication(ctx context.Context, key string) error {
	return c.doJSON(ctx, "DELETE", fmt.Sprintf("/rest-api/v1/applications/%s", pathEscape(key)), nil, nil)
}

func (c *Client) listApplicationGroups(ctx context.Context) ([]applicationGroupBody, error) {
	var out tokenPagedResponse[applicationGroupBody]
	if err := c.doJSON(ctx, "GET", "/rest-api/v1/applicationGroups?pageSize=1000", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) getApplicationGroup(ctx context.Context, key string) (*applicationGroupBody, error) {
	var out applicationGroupBody
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/rest-api/v1/applicationGroups/%s", pathEscape(key)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) createApplicationGroup(ctx context.Context, body applicationGroupBody) (*applicationGroupBody, error) {
	var out applicationGroupBody
	if err := c.doJSON(ctx, "POST", "/rest-api/v1/applicationGroups", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) updateApplicationGroup(ctx context.Context, key string, body applicationGroupBody) (*applicationGroupBody, error) {
	var out applicationGroupBody
	if err := c.doJSON(ctx, "PUT", fmt.Sprintf("/rest-api/v1/applicationGroups/%s", pathEscape(key)), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) deleteApplicationGroup(ctx context.Context, key string) error {
	return c.doJSON(ctx, "DELETE", fmt.Sprintf("/rest-api/v1/applicationGroups/%s", pathEscape(key)), nil, nil)
}

func (c *Client) listTeams(ctx context.Context) ([]orgTeamBody, error) {
	const pageSize = 100
	all := make([]orgTeamBody, 0)
	skip := 0
	for {
		endpoint := fmt.Sprintf("/rest-api/v1/teams?skip=%d&pageSize=%d", skip, pageSize)
		var out apiPagedResponse[orgTeamBody]
		if err := c.doJSON(ctx, "GET", endpoint, nil, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Items...)
		skip += len(out.Items)
		if len(out.Items) == 0 || skip >= out.Paging.TotalItemCount {
			break
		}
	}
	return all, nil
}

func (c *Client) getTeam(ctx context.Context, key string) (*orgTeamBody, error) {
	var out orgTeamBody
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/rest-api/v1/teams/%s", pathEscape(key)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) createTeam(ctx context.Context, body orgTeamBody) (string, error) {
	var out string
	if err := c.doJSON(ctx, "POST", "/rest-api/v1/teams", body, &out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Client) updateTeam(ctx context.Context, key string, body orgTeamBody) (string, error) {
	var out string
	if err := c.doJSON(ctx, "PUT", fmt.Sprintf("/rest-api/v1/teams/%s", pathEscape(key)), body, &out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Client) deleteTeam(ctx context.Context, key string) error {
	return c.doJSON(ctx, "DELETE", fmt.Sprintf("/rest-api/v1/teams/%s", pathEscape(key)), nil, nil)
}

func (c *Client) listRoles(ctx context.Context) ([]roleListItem, error) {
	const pageSize = 100
	all := make([]roleListItem, 0)
	skip := 0
	for {
		endpoint := fmt.Sprintf("/rest-api/v1/roles?skip=%d&pageSize=%d", skip, pageSize)
		var out apiPagedResponse[roleListItem]
		if err := c.doJSON(ctx, "GET", endpoint, nil, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Items...)
		skip += len(out.Items)
		if len(out.Items) == 0 || skip >= out.Paging.TotalItemCount {
			break
		}
	}
	return all, nil
}

func (c *Client) getRole(ctx context.Context, key string) (*roleBody, error) {
	var out roleBody
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/rest-api/v1/roles/%s", pathEscape(key)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) createRole(ctx context.Context, body roleBody) (string, error) {
	var out string
	if err := c.doJSON(ctx, "POST", "/rest-api/v1/roles", body, &out); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Client) deleteRole(ctx context.Context, key string) error {
	return c.doJSON(ctx, "DELETE", fmt.Sprintf("/rest-api/v1/roles/%s", pathEscape(key)), nil, nil)
}

func (c *Client) listEngagements(ctx context.Context) ([]engagementBody, error) {
	var out tokenPagedResponse[engagementBody]
	if err := c.doJSON(ctx, "GET", "/rest-api/v1/engagements?pageSize=1000", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) getEngagement(ctx context.Context, key string) (*engagementBody, error) {
	var out engagementBody
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/rest-api/v1/engagements/%s", pathEscape(key)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) createEngagement(ctx context.Context, body engagementBody) (*engagementBody, error) {
	var out engagementBody
	if err := c.doJSON(ctx, "POST", "/rest-api/v1/engagements", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) updateEngagement(ctx context.Context, key string, body engagementBody) (*engagementBody, error) {
	var out engagementBody
	if err := c.doJSON(ctx, "PUT", fmt.Sprintf("/rest-api/v1/engagements/%s", pathEscape(key)), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) listApplicationProfiles(ctx context.Context) ([]applicationProfileBody, error) {
	var out tokenPagedResponse[applicationProfileBody]
	if err := c.doJSON(ctx, "GET", "/rest-api/v1/applications/profiles?pageSize=1000", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) getApplicationProfile(ctx context.Context, key string) (*applicationProfileBody, error) {
	var out applicationProfileBody
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/rest-api/v1/applications/profiles/%s", pathEscape(key)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) listAuditLogs(ctx context.Context) ([]auditLogBody, error) {
	var out tokenPagedResponse[auditLogBody]
	if err := c.doJSON(ctx, "GET", "/rest-api/v1/auditLogs?pageSize=1000", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) listConnectors(ctx context.Context) ([]connectorBody, error) {
	var out tokenPagedResponse[connectorBody]
	if err := c.doJSON(ctx, "GET", "/rest-api/v1/connectors?pageSize=1000", nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) getConnector(ctx context.Context, id string) (*connectorBody, error) {
	var out connectorBody
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/rest-api/v1/connectors/%s", pathEscape(id)), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
