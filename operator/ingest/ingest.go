package ingest

/*
 Copyright 2018 Crunchy Data Solutions, Inc.
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

import (
	log "github.com/Sirupsen/logrus"
	crv1 "github.com/crunchydata/postgres-operator/apis/cr/v1"
	"github.com/crunchydata/postgres-operator/operator"
	//"github.com/crunchydata/postgres-operator/util"
	"io/ioutil"
	//v1batch "k8s.io/api/batch/v1"
	"k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"bytes"
	"encoding/json"
	"k8s.io/client-go/kubernetes"
	"text/template"
)

type ingestTemplateFields struct {
	Name            string
	PvcName         string
	SecurityContext string
	Namespace       string
	WatchDir        string
	DBHost          string
	DBPort          string
	DBName          string
	DBSecret        string
	DBTable         string
	DBColumn        string
	COImageTag      string
	COImagePrefix   string
	MaxJobs         int
}

const ingestPath = "/operator-conf/pgo-ingest-watch-job.json"

var ingestjobTemplate *template.Template

func init() {
	var err error
	var buf []byte

	buf, err = ioutil.ReadFile(ingestPath)
	if err != nil {
		log.Error(err)
		panic(err.Error())
	}
	ingestjobTemplate = template.Must(template.New("ingest template").Parse(string(buf)))

}

// CreateIngest ...
func CreateIngest(namespace string, clientset *kubernetes.Clientset, client *rest.RESTClient, i *crv1.Pgingest) {

	//create the ingest deployment

	jobFields := ingestTemplateFields{
		Name:            i.Spec.Name,
		PvcName:         i.Spec.PVCName,
		SecurityContext: "",
		Namespace:       namespace,
		WatchDir:        i.Spec.WatchDir,
		DBHost:          i.Spec.DBHost,
		DBPort:          i.Spec.DBPort,
		DBName:          i.Spec.DBName,
		DBSecret:        i.Spec.DBSecret,
		DBTable:         i.Spec.DBTable,
		DBColumn:        i.Spec.DBColumn,
		MaxJobs:         i.Spec.MaxJobs,
		COImageTag:      operator.COImageTag,
		COImagePrefix:   operator.COImagePrefix,
	}

	var doc2 bytes.Buffer
	err := ingestjobTemplate.Execute(&doc2, jobFields)
	if err != nil {
		log.Error(err.Error())
		return
	}
	deploymentDocString := doc2.String()
	log.Debug(deploymentDocString)

	deployment := v1beta1.Deployment{}
	err = json.Unmarshal(doc2.Bytes(), &deployment)
	if err != nil {
		log.Error("error unmarshalling ingest json into Deployment " + err.Error())
		return
	}

	deploymentResult, err := clientset.ExtensionsV1beta1().Deployments(namespace).Create(&deployment)
	if err != nil {
		log.Error("error creating ingest Deployment " + err.Error())
		return
	}

	log.Info("created ingestDeployment " + deploymentResult.Name)

}

// Delete ingest
func Delete(clientset *kubernetes.Clientset, name string, namespace string) error {
	log.Debug("in ingest.Delete")
	var err error

	delOptions := meta_v1.DeleteOptions{}
	var delProp meta_v1.DeletionPropagation
	delProp = meta_v1.DeletePropagationForeground
	delOptions.PropagationPolicy = &delProp

	err = clientset.ExtensionsV1beta1().Deployments(namespace).Delete(name, &delOptions)
	if err != nil {
		log.Error("error deleting replica Deployment " + err.Error())
	}

	return nil

}
