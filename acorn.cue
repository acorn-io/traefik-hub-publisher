containers: default: {
	build: "."
	permissions: {
		clusterRules: [
			{
				verbs: ["*"]
				apiGroups: ["networking.k8s.io"]
				resources: ["ingresses"]
			},
			{
				verbs: ["*"]
				apiGroups: ["hub.traefik.io"]
				resources: ["edgeingresses"]
			},
			{
				verbs: ["get"]
				apiGroups: ["apiextensions.k8s.io"]
				resources: ["customresourcedefinitions"]
				resourceNames: ["edgeingresses.hub.traefik.io"]
			},
		]
	}
}
