package internal

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func getOwners(u unstructured.Unstructured) []string {
	owners := []string{}
	for _, o := range getList(u.Object, "metadata", "ownerReferences") {
		if o == nil {
			continue
		}
		owner, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		uid := getString(owner, "uid")
		owners = append(owners, uid)
	}

	return owners
}

func getMap(m map[string]interface{}, path ...string) map[string]interface{} {
	for _, branch := range path {
		next, ok := m[branch]
		if !ok {
			return nil
		}
		m, ok = next.(map[string]interface{})
		if !ok {
			return nil
		}
	}

	return m
}

func getString(m map[string]interface{}, path ...string) string {
	if len(path) == 0 {
		return ""
	}
	if len(path) > 1 {
		m = getMap(m, path[:len(path)-1]...)
		if m == nil {
			return ""
		}
	}
	leaf, ok := m[path[len(path)-1]]
	if !ok {
		return ""
	}
	val, ok := leaf.(string)
	if !ok {
		return ""
	}
	return val
}

func getList(m map[string]interface{}, path ...string) []interface{} {
	if len(path) == 0 {
		return nil
	}
	if len(path) > 1 {
		m = getMap(m, path[:len(path)-1]...)
		if m == nil {
			return nil
		}
	}
	leaf, ok := m[path[len(path)-1]]
	if !ok {
		return nil
	}
	list, ok := leaf.([]interface{})
	if !ok {
		return nil
	}
	return list
}
