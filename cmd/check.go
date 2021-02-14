package cmd

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "check for broken kube-controller-manager deployment and restart if needed",
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkNRestart(args)
		},
	}
	kcmErrs map[string]int
)

func init() {
	viper.BindPFlags(checkCmd.Flags())
}

func checkNRestart(args []string) error {

	klog.Infoln("Starting kcm-watchdog")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	kcmErrs = make(map[string]int)

	ticker := time.NewTicker(time.Duration(viper.GetDuration("checkinterval")))

	for ; true; <-ticker.C {
		err := checkKCMs(c)
		if err != nil {
			return err
		}
		err = restartKCMs(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkKCMs(c *kubernetes.Clientset) error {
	namespaces, err := c.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ns := range namespaces.Items {
		kcm, err := c.AppsV1().Deployments(ns.Name).Get(context.Background(), "kube-controller-manager", v1.GetOptions{})
		if errors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		if kcm.Status.UnavailableReplicas > 0 {
			kcmErrs[ns.Name]++
			klog.Warningf("kube-controller-manager of %s has unavailable replicas (count %d)", ns.Name, kcmErrs[ns.Name])
		} else {
			kcmErrs[ns.Name] = 0
		}
	}
	return nil
}

func restartKCMs(c *kubernetes.Clientset) error {
	for ns, cn := range kcmErrs {
		if cn >= viper.GetInt("kcm-max-fails") {
			d := c.AppsV1().Deployments(ns)
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				result, err := d.Get(context.Background(), "kube-controller-manager", v1.GetOptions{})
				if err != nil {
					klog.Errorf("Failed to get latest version of kube-controller-manager in namespace %s: %s", ns, err)
				}

				result.Spec.Template.ObjectMeta.Labels["kcm-watchdog-restarted"] = fmt.Sprintf("%d", time.Now().Unix())
				_, err = d.Update(context.Background(), result, v1.UpdateOptions{})
				return err
			})
			if err != nil {
				return err
			}
			klog.Infof("kube-controller-manager in namespace %q restarted", ns)
			kcmErrs[ns] = 0
		}
	}
	return nil
}
