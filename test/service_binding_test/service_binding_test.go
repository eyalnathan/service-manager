/*
 *    Copyright 2018 The Service Manager Authors
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package service_binding_test

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gavv/httpexpect"
	"github.com/gofrs/uuid"

	"testing"

	"github.com/Peripli/service-manager/pkg/types"

	"github.com/Peripli/service-manager/pkg/web"
	"github.com/Peripli/service-manager/test/common"

	"github.com/Peripli/service-manager/test"

	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
)

func TestServiceBindings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Bindings Tests Suite")
}

const (
	TenantIdentifier = "tenant"
	TenantIDValue    = "tenantID"
)

var _ = test.DescribeTestsFor(test.TestCase{
	API: web.ServiceBindingsURL,
	SupportedOps: []test.Op{
		test.Get, test.List, test.Delete,
	},
	MultitenancySettings: &test.MultitenancySettings{
		ClientID:           "tenancyClient",
		ClientIDTokenClaim: "cid",
		TenantTokenClaim:   "zid",
		LabelKey:           TenantIdentifier,
		TokenClaims: map[string]interface{}{
			"cid": "tenancyClient",
			"zid": TenantIDValue,
		},
	},
	ResourceType:                           types.ServiceBindingType,
	SupportsAsyncOperations:                true,
	DisableTenantResources:                 false,
	StrictlyTenantScoped:                   true,
	ResourceBlueprint:                      blueprint,
	ResourceWithoutNullableFieldsBlueprint: blueprint,
	ResourcePropertiesToIgnore:             []string{"volume_mounts", "endpoints", "bind_resource", "credentials"},
	PatchResource:                          test.StorageResourcePatch,
	AdditionalTests: func(ctx *common.TestContext) {
		Context("additional non-generic tests", func() {
			var (
				postBindingRequest      common.Object
				expectedBindingResponse common.Object

				smExpect *common.SMExpect
			)

			createInstance := func(SM *common.SMExpect) string {
				instanceID, err := uuid.NewV4()
				if err != nil {
					panic(err)
				}

				planID := newServicePlan(ctx)
				test.EnsurePlanVisibility(ctx.SMRepository, TenantIdentifier, types.SMPlatform, planID, TenantIDValue)

				instanceBody := common.Object{
					"id":               instanceID.String(),
					"name":             "test-instance",
					"service_plan_id":  planID,
					"maintenance_info": "{}",
				}

				SM.POST(web.ServiceInstancesURL).WithJSON(instanceBody).
					Expect().
					Status(http.StatusCreated)

				return instanceID.String()
			}

			createBinding := func(SM *common.SMExpect, body common.Object) {
				SM.POST(web.ServiceBindingsURL).WithJSON(body).
					Expect().
					Status(http.StatusCreated).
					JSON().Object().
					ContainsMap(expectedBindingResponse).ContainsKey("id")
			}

			BeforeEach(func() {
				smExpect = ctx.SMWithOAuthForTenant // by default all requests are tenant-scoped
			})

			JustBeforeEach(func() {
				var err error
				instanceID := createInstance(smExpect)

				bindingID, err := uuid.NewV4()
				if err != nil {
					panic(err)
				}

				bindingName := "test-binding"

				postBindingRequest = common.Object{
					"id":                  bindingID.String(),
					"name":                bindingName,
					"service_instance_id": instanceID,
				}
				expectedBindingResponse = common.Object{
					"id":                  bindingID.String(),
					"name":                bindingName,
					"service_instance_id": instanceID,
				}
			})

			AfterEach(func() {
				ctx.CleanupAdditionalResources()
			})

			Describe("POST", func() {
				Context("when content type is not JSON", func() {
					It("returns 415", func() {
						smExpect.POST(web.ServiceBindingsURL).WithText("text").
							Expect().
							Status(http.StatusUnsupportedMediaType).
							JSON().Object().
							Keys().Contains("error", "description")
					})
				})

				Context("when request body is not a valid JSON", func() {
					It("returns 400", func() {
						smExpect.POST(web.ServiceBindingsURL).
							WithText("invalid json").
							WithHeader("content-type", "application/json").
							Expect().
							Status(http.StatusBadRequest).
							JSON().Object().
							Keys().Contains("error", "description")
					})
				})

				Context("when a request body field is missing", func() {
					assertPOSTReturns400WhenFieldIsMissing := func(field string) {
						JustBeforeEach(func() {
							delete(postBindingRequest, field)
							delete(expectedBindingResponse, field)
						})

						It("returns 400", func() {
							smExpect.POST(web.ServiceBindingsURL).WithJSON(postBindingRequest).
								Expect().
								Status(http.StatusBadRequest).
								JSON().Object().
								Keys().Contains("error", "description")
						})
					}

					assertPOSTReturns201WhenFieldIsMissing := func(field string) {
						JustBeforeEach(func() {
							delete(postBindingRequest, field)
							delete(expectedBindingResponse, field)
						})

						It("returns 201", func() {
							createBinding(smExpect, postBindingRequest)
						})
					}

					Context("when id  field is missing", func() {
						assertPOSTReturns201WhenFieldIsMissing("id")
					})

					Context("when name field is missing", func() {
						assertPOSTReturns400WhenFieldIsMissing("name")
					})

					Context("when service_instance_id field is missing", func() {
						assertPOSTReturns400WhenFieldIsMissing("service_instance_id")
					})

				})

				Context("when request body id field is invalid", func() {
					It("should return 400", func() {
						postBindingRequest["id"] = "binding/1"
						resp := smExpect.POST(web.ServiceBindingsURL).
							WithJSON(postBindingRequest).
							Expect().Status(http.StatusBadRequest).JSON().Object()

						resp.Value("description").Equal("binding/1 contains invalid character(s)")
					})
				})

				Context("With async query param", func() {
					It("succeeds", func() {
						resp := smExpect.POST(web.ServiceBindingsURL).WithJSON(postBindingRequest).
							WithQuery("async", "true").
							Expect().
							Status(http.StatusAccepted)

						test.ExpectOperation(smExpect, resp, types.SUCCEEDED)

						smExpect.GET(fmt.Sprintf("%s/%s", web.ServiceBindingsURL, postBindingRequest["id"])).Expect().
							Status(http.StatusOK).
							JSON().Object().
							ContainsMap(expectedBindingResponse).ContainsKey("id")
					})
				})

				Context("instance ownership", func() {
					When("no tenant is specified", func() {
						It("returns 400", func() {
							ctx.SMWithOAuth.POST(web.ServiceBindingsURL).
								WithJSON(postBindingRequest).
								Expect().Status(http.StatusBadRequest)
						})
					})

					When("tenant has ownership of instance", func() {
						It("returns 201", func() {
							smExpect.POST(web.ServiceBindingsURL).
								WithJSON(postBindingRequest).
								Expect().Status(http.StatusCreated)
						})
					})
				})
			})

			Describe("DELETE", func() {
				Context("instance ownership", func() {
					When("tenant doesn't have ownership of binding", func() {
						var smWithOtherTenant *httpexpect.Expect
						BeforeEach(func() {
							oauthServer := ctx.Servers[common.OauthServer].(*common.OAuthServer)
							accessToken := oauthServer.CreateToken(map[string]interface{}{
								"cid": "tenancyClient",
								"zid": "otherTenant",
							})
							smWithOtherTenant = ctx.SM.Builder(func(req *httpexpect.Request) {
								req.WithHeader("Authorization", "Bearer "+accessToken)
							})
						})

						It("returns 404", func() {
							ctx.SMWithOAuthForTenant.POST(web.ServiceBindingsURL).WithJSON(postBindingRequest).
								Expect().
								Status(http.StatusCreated)

							smWithOtherTenant.DELETE(fmt.Sprintf("%s/%s", web.ServiceBindingsURL, postBindingRequest["id"])).
								Expect().Status(http.StatusNotFound)
						})
					})

					When("tenant has ownership of instance", func() {
						BeforeEach(func() {
							smExpect = ctx.SMWithOAuthForTenant
						})

						It("returns 200", func() {
							smExpect.POST(web.ServiceBindingsURL).WithJSON(postBindingRequest).
								Expect().
								Status(http.StatusCreated)

							smExpect.DELETE(fmt.Sprintf("%s/%s", web.ServiceBindingsURL, postBindingRequest["id"])).
								Expect().Status(http.StatusOK)
						})
					})
				})
			})

		})
	},
})

func blueprint(ctx *common.TestContext, auth *common.SMExpect, async bool) common.Object {
	ID, err := uuid.NewV4()
	if err != nil {
		Fail(fmt.Sprintf("failed to generate instance GUID: %s", err))
	}
	instanceID := "instance-" + ID.String()

	servicePlanID := newServicePlan(ctx)
	test.EnsurePlanVisibility(ctx.SMRepository, TenantIdentifier, types.SMPlatform, servicePlanID, TenantIDValue)
	resp := ctx.SMWithOAuthForTenant.POST(web.ServiceInstancesURL).
		WithQuery("async", strconv.FormatBool(async)).
		WithJSON(common.Object{
			"id":               instanceID,
			"name":             instanceID + "name",
			"service_plan_id":  servicePlanID,
			"maintenance_info": "{}",
		}).Expect()

	if async {
		test.ExpectSuccessfulAsyncResourceCreation(resp, auth, instanceID, web.ServiceInstancesURL)
	} else {
		resp.Status(http.StatusCreated)
	}

	bindingID := "binding-" + ID.String()

	resp = ctx.SMWithOAuthForTenant.POST(web.ServiceBindingsURL).
		WithQuery("async", strconv.FormatBool(async)).
		WithJSON(common.Object{
			"id":                  bindingID,
			"name":                bindingID + "name",
			"service_instance_id": instanceID,
		}).Expect()

	var binding map[string]interface{}
	if async {
		binding = test.ExpectSuccessfulAsyncResourceCreation(resp, auth, bindingID, web.ServiceBindingsURL)
	} else {
		binding = resp.Status(http.StatusCreated).JSON().Object().Raw()
	}

	delete(binding, "credentials")
	return binding
}

func newServicePlan(ctx *common.TestContext) string {
	brokerID, _, _ := ctx.RegisterBrokerWithCatalog(common.NewRandomSBCatalog())
	so := ctx.SMWithOAuth.ListWithQuery(web.ServiceOfferingsURL, fmt.Sprintf("fieldQuery=broker_id eq '%s'", brokerID)).First()
	servicePlanID := ctx.SMWithOAuth.ListWithQuery(web.ServicePlansURL, "fieldQuery="+fmt.Sprintf("service_offering_id eq '%s'", so.Object().Value("id").String().Raw())).
		First().Object().Value("id").String().Raw()
	return servicePlanID
}