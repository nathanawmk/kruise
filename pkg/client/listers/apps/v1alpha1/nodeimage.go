/*
Copyright 2020 The Kruise Authors.

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
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/openkruise/kruise/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// NodeImageLister helps list NodeImages.
type NodeImageLister interface {
	// List lists all NodeImages in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.NodeImage, err error)
	// NodeImages returns an object that can list and get NodeImages.
	NodeImages(namespace string) NodeImageNamespaceLister
	NodeImageListerExpansion
}

// nodeImageLister implements the NodeImageLister interface.
type nodeImageLister struct {
	indexer cache.Indexer
}

// NewNodeImageLister returns a new NodeImageLister.
func NewNodeImageLister(indexer cache.Indexer) NodeImageLister {
	return &nodeImageLister{indexer: indexer}
}

// List lists all NodeImages in the indexer.
func (s *nodeImageLister) List(selector labels.Selector) (ret []*v1alpha1.NodeImage, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.NodeImage))
	})
	return ret, err
}

// NodeImages returns an object that can list and get NodeImages.
func (s *nodeImageLister) NodeImages(namespace string) NodeImageNamespaceLister {
	return nodeImageNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// NodeImageNamespaceLister helps list and get NodeImages.
type NodeImageNamespaceLister interface {
	// List lists all NodeImages in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.NodeImage, err error)
	// Get retrieves the NodeImage from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.NodeImage, error)
	NodeImageNamespaceListerExpansion
}

// nodeImageNamespaceLister implements the NodeImageNamespaceLister
// interface.
type nodeImageNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all NodeImages in the indexer for a given namespace.
func (s nodeImageNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.NodeImage, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.NodeImage))
	})
	return ret, err
}

// Get retrieves the NodeImage from the indexer for a given namespace and name.
func (s nodeImageNamespaceLister) Get(name string) (*v1alpha1.NodeImage, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("nodeimage"), name)
	}
	return obj.(*v1alpha1.NodeImage), nil
}
