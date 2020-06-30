package advertisement_operator

import (
	"context"
	protocolv1 "github.com/liqoTech/liqo/api/advertisement-operator/v1"
	policyv1 "github.com/liqoTech/liqo/api/cluster-config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
)

func (b *AdvertisementBroadcaster) WatchConfiguration(kubeconfigPath string) error {
	configClient, err := policyv1.CreateClusterConfigClient(kubeconfigPath)
	if err != nil {
		klog.Error(err)
		return err
	}

	watcher, err := configClient.Resource("clusterconfigs").Watch(metav1.ListOptions{})
	if err != nil {
		return err
	}

	go func() {
		for event := range watcher.ResultChan() {
			configuration, ok := event.Object.(*policyv1.ClusterConfig)
			if !ok {
				continue
			}

			switch event.Type {
			case watch.Added, watch.Modified:
				b.ClusterConfig.ResourceSharingPercentage = configuration.Spec.ResourceSharingPercentage
			case watch.Deleted:
				// TODO: set default config?
			}
		}
	}()
	return nil
}

func (r *AdvertisementReconciler) WatchConfiguration(kubeconfigPath string) error {
	configClient, err := policyv1.CreateClusterConfigClient(kubeconfigPath)
	if err != nil {
		klog.Error(err)
		return err
	}

	watcher, err := configClient.Resource("clusterconfigs").Watch(metav1.ListOptions{})
	if err != nil {
		return err
	}

	go func() {
		for event := range watcher.ResultChan() {
			configuration, ok := event.Object.(*policyv1.ClusterConfig)
			if !ok {
				continue
			}

			switch event.Type {
			case watch.Added, watch.Modified:
				obj, err := r.AdvClient.Resource("advertisements").List(metav1.ListOptions{})
				if err != nil {
					klog.Errorln(err, "Unable to apply configuration: error listing Advertisements")
					continue
				}
				advList := obj.(*protocolv1.AdvertisementList)
				err, updateFlag := r.ManageConfigUpdate(configuration, advList)
				if err != nil {
					continue
				}
				if updateFlag {
					for _, adv := range advList.Items {
						r.UpdateAdvertisement(&adv)
					}
				}
			case watch.Deleted:
				// TODO: set default config?
			}
		}
	}()
	return nil
}

func (r *AdvertisementReconciler) ManageConfigUpdate(configuration *policyv1.ClusterConfig, advList *protocolv1.AdvertisementList) (error, bool) {

	updateFlag := false
	if configuration.Spec.MaxAcceptableAdvertisement > r.ClusterConfig.MaxAcceptableAdvertisement {
		// the maximum has increased: check if there are refused advertisements which now can be accepted
		r.ClusterConfig = configuration.Spec
		for i := 0; i < len(advList.Items); i++ {
			adv := &advList.Items[i]
			if adv.Status.AdvertisementStatus == "REFUSED" {
				r.CheckAdvertisement(adv)
				updateFlag = true
			}
		}
	} else {
		// the maximum has decreased: if the already accepted advertisements are too many (with the new maximum), delete some of them
		r.ClusterConfig = configuration.Spec
		if r.ClusterConfig.MaxAcceptableAdvertisement < r.AcceptedAdvNum {
			for i := 0; i < int(r.AcceptedAdvNum-r.ClusterConfig.MaxAcceptableAdvertisement); i++ {
				adv := advList.Items[i]
				if adv.Status.AdvertisementStatus == "ACCEPTED" {
					err := r.Client.Delete(context.Background(), &adv)
					if err != nil {
						klog.Errorln(err, "Unable to apply configuration: error deleting Advertisement "+adv.Name)
						return err, updateFlag
					}
					r.AcceptedAdvNum--
				}
			}
		}
	}
	return nil, updateFlag
}
