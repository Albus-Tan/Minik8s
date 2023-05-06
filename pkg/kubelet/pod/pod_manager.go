package pod

import (
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"sync"
)

// BuildPodFullName builds the pod full name from pod name and namespace.
func BuildPodFullName(name, namespace string) string {
	return name + "_" + namespace
}

// GetPodFullName returns a name that uniquely identifies a pod.
func GetPodFullName(pod *core.Pod) string {
	return pod.Name + "_" + pod.Namespace
}

// Manager stores and manages access to pods
type Manager interface {
	// GetPods returns the regular pods bound to the kubelet and their spec.
	GetPods() []*core.Pod
	// GetPodByName provides the pod that matches namespace and
	// name, as well as whether the pod was found.
	GetPodByName(namespace, name string) (*core.Pod, bool)
	// GetPodByUID provides the pod that matches pod UID, as well as
	// whether the pod is found.
	GetPodByUID(types.UID) (*core.Pod, bool)
	// AddPod adds the given pod to the manager.
	AddPod(pod *core.Pod)
	// UpdatePod updates the given pod in the manager.
	UpdatePod(pod *core.Pod)
	// DeletePod deletes the given pod from the manager.
	DeletePod(pod *core.Pod)
}

// basicManager is a functional Manager.
//
// All fields in basicManager are read-only and are updated calling SetPods,
// AddPod, UpdatePod, or DeletePod.
type basicManager struct {
	// Protects all internal maps.
	lock sync.RWMutex

	// Regular pods indexed by UID.
	podByUID map[types.UID]*core.Pod

	// Pods indexed by full name for easy access.
	podByFullName map[string]*core.Pod
}

// NewPodManager returns a functional Manager.
func NewPodManager() Manager {
	pm := &basicManager{}

	pm.lock.Lock()
	defer pm.lock.Unlock()

	pm.podByUID = make(map[types.UID]*core.Pod)
	pm.podByFullName = make(map[string]*core.Pod)

	return pm
}

func (b *basicManager) GetPods() []*core.Pod {
	b.lock.RLock()
	defer b.lock.RUnlock()

	pods := make([]*core.Pod, 0, len(b.podByUID))
	for _, pod := range b.podByUID {
		pods = append(pods, pod)
	}
	return pods
}

func (b *basicManager) GetPodByName(namespace, name string) (*core.Pod, bool) {
	podFullName := BuildPodFullName(name, namespace)
	b.lock.RLock()
	defer b.lock.RUnlock()
	pod, ok := b.podByFullName[podFullName]
	return pod, ok
}

func (b *basicManager) GetPodByUID(uid types.UID) (*core.Pod, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	pod, ok := b.podByUID[uid]
	return pod, ok
}

func (b *basicManager) AddPod(pod *core.Pod) {
	b.UpdatePod(pod)
}

func (b *basicManager) UpdatePod(pod *core.Pod) {
	b.lock.Lock()
	defer b.lock.Unlock()

	podFullName := GetPodFullName(pod)
	podUID := pod.UID

	b.podByUID[podUID] = pod
	b.podByFullName[podFullName] = pod
}

func (b *basicManager) DeletePod(pod *core.Pod) {
	b.lock.Lock()
	defer b.lock.Unlock()
	podFullName := GetPodFullName(pod)
	podUID := pod.UID
	delete(b.podByUID, podUID)
	delete(b.podByFullName, podFullName)
}
