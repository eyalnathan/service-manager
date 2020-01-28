package interceptors

import (
	"context"
	"fmt"
	"github.com/Peripli/service-manager/pkg/query"
	"github.com/Peripli/service-manager/pkg/types"
	"github.com/Peripli/service-manager/storage"
)

func NewBrokerNotificationsInterceptor() *NotificationsInterceptor {
	return &NotificationsInterceptor{
		PlatformIDsProviderFunc: func(ctx context.Context, obj types.Object, repository storage.Repository) ([]string, error) {
			broker := obj.(*types.ServiceBroker)

			var err error
			planIDs := make([]string, 0)
			if len(broker.Services) == 0 {
				planIDs, err = fetchBrokerPlanIDs(ctx, broker.ID, repository)
				if err != nil {
					return nil, err
				}
			} else {
				for _, svc := range broker.Services {
					for _, plan := range svc.Plans {
						planIDs = append(planIDs, plan.ID)
					}
				}
			}

			if len(planIDs) == 0 {
				return []string{}, nil
			}

			byPlanIDs := query.ByField(query.InOperator, "service_plan_id", planIDs...)
			objList, err := repository.List(ctx, types.VisibilityType, byPlanIDs)
			if err != nil {
				return nil, err
			}

			hasPublicPlan := false
			platformIDs := make([]string, 0)
			for i := 0; i < objList.Len(); i++ {
				platformID := objList.ItemAt(i).(*types.Visibility).PlatformID
				if platformID == "" {
					hasPublicPlan = true
					break
				}

				platformIDs = append(platformIDs, platformID)
			}

			if hasPublicPlan {
				objList, err := repository.List(ctx, types.PlatformType, query.ByField(query.NotEqualsOperator, "id", types.SMPlatform))
				if err != nil {
					return nil, err
				}

				platformIDs = make([]string, 0)
				for i := 0; i < objList.Len(); i++ {
					platformIDs = append(platformIDs, objList.ItemAt(i).(*types.Platform).ID)
				}
			}

			return platformIDs, nil
		},
		AdditionalDetailsFunc: func(ctx context.Context, objects types.ObjectList, repository storage.Repository) (objectDetails, error) {
			details := make(objectDetails, objects.Len())
			for i := 0; i < objects.Len(); i++ {
				broker := objects.ItemAt(i).(*types.ServiceBroker)
				details[broker.ID] = &BrokerAdditional{
					Services: broker.Services,
				}
			}
			return details, nil
		},
	}
}

type BrokerAdditional struct {
	Services []*types.ServiceOffering `json:"services,omitempty"`
}

func (ba BrokerAdditional) Validate() error {
	if len(ba.Services) == 0 {
		return fmt.Errorf("broker details services cannot be empty")
	}

	return nil
}

const (
	BrokerCreateNotificationInterceptorName = "BrokerNotificationsCreateInterceptorProvider"
	BrokerUpdateNotificationInterceptorName = "BrokerNotificationsUpdateInterceptorProvider"
	BrokerDeleteNotificationInterceptorName = "BrokerNotificationsDeleteInterceptorProvider"
)

type BrokerNotificationsCreateInterceptorProvider struct {
}

func (*BrokerNotificationsCreateInterceptorProvider) Name() string {
	return BrokerCreateNotificationInterceptorName
}

func (*BrokerNotificationsCreateInterceptorProvider) Provide() storage.CreateOnTxInterceptor {
	return NewBrokerNotificationsInterceptor()
}

type BrokerNotificationsUpdateInterceptorProvider struct {
}

func (*BrokerNotificationsUpdateInterceptorProvider) Name() string {
	return BrokerUpdateNotificationInterceptorName
}

func (*BrokerNotificationsUpdateInterceptorProvider) Provide() storage.UpdateOnTxInterceptor {
	return NewBrokerNotificationsInterceptor()
}

type BrokerNotificationsDeleteInterceptorProvider struct {
}

func (*BrokerNotificationsDeleteInterceptorProvider) Name() string {
	return BrokerDeleteNotificationInterceptorName
}

func (*BrokerNotificationsDeleteInterceptorProvider) Provide() storage.DeleteOnTxInterceptor {
	return NewBrokerNotificationsInterceptor()
}

func fetchBrokerPlanIDs(ctx context.Context, brokerID string, repository storage.Repository) ([]string, error) {
	byBrokerID := query.ByField(query.EqualsOperator, "broker_id", brokerID)
	objList, err := repository.List(ctx, types.ServiceOfferingType, byBrokerID)
	if err != nil {
		return nil, err
	}

	if objList.Len() == 0 {
		return []string{}, nil
	}

	serviceOfferingIDs := make([]string, 0)
	for i := 0; i < objList.Len(); i++ {
		serviceOfferingIDs = append(serviceOfferingIDs, objList.ItemAt(i).GetID())
	}

	byOfferingIDs := query.ByField(query.InOperator, "service_offering_id", serviceOfferingIDs...)
	objList, err = repository.List(ctx, types.ServicePlanType, byOfferingIDs)
	if err != nil {
		return nil, err
	}

	planIDs := make([]string, 0)
	for i := 0; i < objList.Len(); i++ {
		planIDs = append(planIDs, objList.ItemAt(i).GetID())
	}

	return planIDs, nil
}
