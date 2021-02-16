package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	checkShoot = &cobra.Command{
		Use:   "check",
		Short: "check for broken shoot deployments and restart if needed",
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkNRestart(args)
		},
	}
	deplErrs map[string]int
)

func init() {
	viper.BindPFlags(checkShoot.Flags())
}

func checkNRestart(args []string) error {

	klog.Infoln("Starting shoot-watchdog")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	deplErrs = make(map[string]int)

	ticker := time.NewTicker(time.Duration(viper.GetDuration("checkinterval")))

	for ; true; <-ticker.C {
		err := checkDeployments(c)
		if err != nil {
			return err
		}
		err = restartDeployments(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkDeployments(c *kubernetes.Clientset) error {
	namespaces, err := c.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, ns := range namespaces.Items {
		if !strings.Contains(ns.Name, "shoot--") {
			continue
		}
		deployments, err := c.AppsV1().Deployments(ns.Name).List(context.Background(), v1.ListOptions{})
		if err != nil {
			return err
		}
		for _, depl := range deployments.Items {
			if depl.Status.UnavailableReplicas > 0 {
				deplErrs[ns.Name+"/"+depl.Name]++
				klog.Warningf("%s of %s has unavailable replicas (count %d)", depl.Name, ns.Name, deplErrs[ns.Name+"/"+depl.Name])
			} else {
				deplErrs[ns.Name+"/"+depl.Name] = 0
			}
		}
	}
	return nil
}

func restartDeployments(c *kubernetes.Clientset) error {
	for i, cn := range deplErrs {
		depls := strings.Split(i, "/")
		if len(depls) != 2 {
			klog.Errorf("something went horrible wrong %s", i)
			continue
		}
		if cn >= viper.GetInt("depl-max-fails") {
			d := c.AppsV1().Deployments(depls[0])
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				result, err := d.Get(context.Background(), depls[1], v1.GetOptions{})
				if err != nil {
					klog.Errorf("Failed to get latest version of %s in namespace %s: %s", depls[1], depls[0], err)
				}

				result.Spec.Template.ObjectMeta.Labels["shoot-watchdog-restarted"] = fmt.Sprintf("%d", time.Now().Unix())
				_, err = d.Update(context.Background(), result, v1.UpdateOptions{})
				return err
			})
			if err != nil {
				return err
			}
			klog.Infof("%q in namespace %q restarted", depls[1], depls[0])
			deplErrs[i] = 0
		}
	}
	return nil
}
