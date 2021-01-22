package connector

import (
	v1 "k8s.io/api/core/v1"

	"github.com/datawire/telepresence2/rpc/manager"
)

func (kc *k8sCluster) detectIngressBehavior() []*manager.IngressInfo {
	loadBalancers := kc.findAllSvcByType(v1.ServiceTypeLoadBalancer)

	type portFilter func(p *v1.ServicePort) bool
	findTCPPort := func(ports []v1.ServicePort, filter portFilter) *v1.ServicePort {
		for i := range ports {
			p := &ports[i]
			if p.Protocol == "" || p.Protocol == "TCP" && filter(p) {
				return p
			}
		}
		return nil
	}

	// filters in priority order.
	portFilters := []portFilter{
		func(p *v1.ServicePort) bool { return p.Name == "https" },
		func(p *v1.ServicePort) bool { return p.Port == 443 },
		func(p *v1.ServicePort) bool { return p.Name == "http" },
		func(p *v1.ServicePort) bool { return p.Port == 80 },
		func(p *v1.ServicePort) bool { return true },
	}

	iis := make([]*manager.IngressInfo, 0, len(loadBalancers))
	for _, lb := range loadBalancers {
		spec := &lb.Spec
		var port *v1.ServicePort
		for _, pf := range portFilters {
			if p := findTCPPort(spec.Ports, pf); p != nil {
				port = p
				break
			}
		}
		if port == nil {
			continue
		}

		iis = append(iis, &manager.IngressInfo{
			Host:   lb.Name + "." + lb.Namespace,
			UseTls: port.Port == 443,
			Port:   port.Port,
		})
	}
	return iis
}
