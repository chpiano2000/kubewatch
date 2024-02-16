/*
Copyright 2016 Skippbox, Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bcr

import (
	"context"
	"fmt"
	"github.com/bizflycloud/gobizfly"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
	"github.com/bitnami-labs/kubewatch/pkg/utils"
)

type BCR struct {
	Name     string
	Username string
	Password string
	Server   string
	Gobizfly *gobizfly.Client
}

// Init prepares slack configuration
func (s *BCR) Init(c *config.Config) error {
	name := c.Handler.Bcr.Name
	username := c.Handler.Bcr.Username
	password := c.Handler.Bcr.Password
	server := c.Handler.Bcr.Server
	auth := utils.BizflyAuth{
		Host:          c.Handler.Bcr.Host,
		AppCredId:     c.Handler.Bcr.AppCredId,
		AppCredSecret: c.Handler.Bcr.AppCredSecret,
		Region:        c.Handler.Bcr.Region,
		BasicAuth:     c.Handler.Bcr.BasicAuth,
	}
	bizflyClient, _ := utils.GetApiClient(&auth)

	s.Name = name
	s.Username = username
	s.Password = password
	s.Server = server
	s.Gobizfly = bizflyClient
	return nil
}

// Handle handles the notification.
func (s *BCR) Handle(e event.Event) {
	// TODO: get bcrIntegrated using gobizfly
	bcrIntegrated := true
	if !bcrIntegrated {
		return
	}
	var kubeClient kubernetes.Interface

	kubeClient = utils.GetClientOutOfCluster()

	switch e.Kind {
	case "Namespace":
		if e.Reason != "Created" {
			return
		}

		secretBody := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      s.Name,
				Namespace: e.Name,
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				corev1.DockerConfigJsonKey: []byte(fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","email":"none"}}}`, s.Server, s.Username, s.Password)),
			},
		}
		_, err := kubeClient.CoreV1().Secrets(e.Name).Create(context.TODO(), secretBody, metav1.CreateOptions{})
		if err != nil {
			logrus.Printf("Create secret error: %s", err.Error())
		}
		logrus.Printf("Secret Created in namespace %s", e.Name)

	case "Secret":
		if e.Reason != "Deleted" {
			return
		}
		if secretObj, ok := e.Obj.(*corev1.Secret); ok {
			_, bcr_secret := secretObj.Data[".dockerconfigjson"]
			if bcr_secret {
				secretBody := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      s.Name,
						Namespace: e.Namespace,
					},
					Type: corev1.SecretTypeDockerConfigJson,
					Data: map[string][]byte{
						corev1.DockerConfigJsonKey: []byte(fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","email":"none"}}}`, s.Server, s.Username, s.Password)),
					},
				}
				_, err := kubeClient.CoreV1().Secrets(e.Namespace).Create(context.TODO(), secretBody, metav1.CreateOptions{})
				if err != nil {
					logrus.Printf("Recreate secret error: %s", err.Error())
				}
				logrus.Printf("Secret Recreated in namespace %s", e.Name)
			} else {
				logrus.Printf("Not a BCR integrated secret, skip processing")
			}
		} else {
			logrus.Printf("Not a Secret Object, skip processing")
		}
	}
}
