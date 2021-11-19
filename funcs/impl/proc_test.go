package impl

import (
	"testing"
)

func TestGetCmdline(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{name: "case0", args: args{command: "/usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --max-pods 64 --pod-max-pids 16384 --pod-manifest-path=/etc/kubernetes/manifests --feature-gates=IPv6DualStack=true --network-plugin=cni --cni-conf-dir=/etc/cni/net.d --cni-bin-dir=/opt/cni/bin --dynamic-config-dir=/etc/kubernetes/kubelet-config --v=3 --enable-controller-attach-detach=true --cluster-dns=192.168.0.10 --pod-infra-container-image=registry-vpc.cn-wulanchabu.aliyuncs.com/acs/pause:3.5 --enable-load-reader --cluster-domain=cluster.local --cloud-provider=external --hostname-override=cn-wulanchabu.172.30.113.137 --provider-id=cn-wulanchabu.i-0jl292pbpfbkntp8o5qq --authorization-mode=Webhook --authentication-token-webhook=true --anonymous-auth=false --client-ca-file=/etc/kubernetes/pki/ca.crt --cgroup-driver=systemd --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_GCM_SHA256 --tls-cert-file=/var/lib/kubelet/pki/kubelet.crt --tls-private-key-file=/var/lib/kubelet/pki/kubelet.key --rotate-certificates=true --cert-dir=/var/lib/kubelet/pki --node-labels=ack.aliyun.com=cce9a2e7d4ffc40a0801844989d633873,ack.aliyun.com=cce9a2e7d4ffc40a0801844989d633873 --system-reserved=cpu=50m,memory=216Mi --kube-reserved=cpu=50m,memory=216Mi --kube-reserved=pid=1000 --system-reserved=pid=1000"}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCmdline(tt.args.command); got == nil {
				t.Errorf("GetCmdline() = %v, want %v", got, tt.want)
			} else {
				for s, val := range got {
					t.Logf("key=%s  val=%s", s, val)
				}
			}
		})
	}
}

func TestGetListeningPorts(t *testing.T) {
	tests := []struct {
		name string
		want []map[string]interface{}
	}{
		{name: "case", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetListeningPorts(); got == nil {
				t.Errorf("GetListeningPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCommandLine(t *testing.T) {
	type args struct {
		processName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "case", args: args{processName: "datakit"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCommandLine(tt.args.processName); got != tt.want {
				t.Errorf("GetCommandLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
