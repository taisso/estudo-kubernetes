# Estudo Kubernetes – Go Server

Aplicação simples em Go para estudar conceitos de Kubernetes: Deployments, Services, ConfigMaps, Secrets, Probes, PVC, HPA, Ingress, RBAC e Metrics Server. Inclui `Dockerfile` e manifests em `k8s/`.

## Sumário
- O que é o projeto
- Endpoints da aplicação
- Pré-requisitos
- Como executar localmente (sem Kubernetes)
- Como usar Docker
- Como criar um cluster com Kind
- Deploy no Kubernetes (manifests em `k8s/`)
- Observabilidade e Autoescalonamento (HPA)
- Notas sobre Ingress e TLS

## O que é o projeto
Um servidor HTTP em Go que expõe rotas para demonstrar leitura de variáveis de ambiente via ConfigMap/Secret, montagem de arquivo via ConfigMap (volume), e health checks para probes.

Código principal: `server.go`
- Porta: 8080
- Variáveis de ambiente usadas:
  - `NAME`, `AGE` (via ConfigMap `goserver-env`)
  - `USER`, `PASSWORD` (via Secret `goserver-secret`)
- Arquivo montado por volume: `/go/myfamily/family.txt` (ConfigMap `configmap-family`)

## Endpoints
- `/` – responde "Hello, I'm ${NAME}. I'm ${AGE}."
- `/config-map` – lê e responde o conteúdo de `/go/myfamily/family.txt`.
- `/secret` – imprime `USER` e `PASSWORD`.
- `/healthz` – retorna 500 nos primeiros ~10s após start; depois retorna 200 OK. Usado por startup/readiness/liveness probes.

Exemplos:
```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/
curl -i http://localhost:8080/config-map
curl -i http://localhost:8080/secret
```

## Pré-requisitos
- Go 1.22+
- Docker 24+
- kubectl 1.27+
- kind (se for usar Kubernetes local)
- Opcional: NGINX Ingress Controller, cert-manager (para Ingress/TLS)

## Executar localmente (sem Kubernetes)
```bash
# dentro do diretório do projeto
export NAME="taisso" AGE="25" USER="luiz" PASSWORD="123456"
go run .
# em outro terminal
curl http://localhost:8080/
```
Obs.: `/config-map` depende do arquivo `/go/myfamily/family.txt`, que só é montado via volume no Kubernetes. Localmente, essa rota retornará erro a menos que você crie o arquivo manualmente.

## Usar Docker
Build e execução local do container:
```bash
docker build -t taisso/hello-go:local .
docker run --rm -p 8080:8080 \
  -e NAME=taisso -e AGE=25 -e USER=luiz -e PASSWORD=123456 \
  taisso/hello-go:local
```

## Criar um cluster com Kind
Há um arquivo de configuração em `k8s/kind.yaml`.
```bash
kind create cluster --name estudo-k8s --config k8s/kind.yaml
kubectl cluster-info --context kind-estudo-k8s
```

Instale o NGINX Ingress Controller (se for usar Ingress):
```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

Opcional: instalar cert-manager (para TLS/Let’s Encrypt):
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.4/cert-manager.yaml
```

## Deploy no Kubernetes
Os manifests estão em `k8s/`. A imagem padrão usada em `k8s/deployment.yaml` é `taisso/hello-go:v5.3`. Ajuste se necessário.

Aplicando recursos principais (namespace default):
```bash
kubectl apply -f k8s/configmap-env.yaml
kubectl apply -f k8s/configmap-family.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/pvc.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

- `Deployment` configura probes de startup/readiness/liveness em `/healthz`.
- `Service` expõe como `LoadBalancer` (em cloud) ou ClusterIP (comente/descomente no YAML conforme necessário).
- `PVC` é montado em `/go/pvc` (exemplo) e o ConfigMap `configmap-family` é montado em `/go/myfamily/family.txt`.

Verifique:
```bash
kubectl get pods,svc,cm,secret,pvc
kubectl logs deploy/goserver-test
kubectl port-forward svc/goserver-service 8080:80
curl http://localhost:8080/
```

### HPA (Horizontal Pod Autoscaler)
Habilite o Metrics Server antes:
```bash
kubectl apply -f k8s/metrics-server.yaml
```

Aplique o HPA:
```bash
kubectl apply -f k8s/hpa.yaml
kubectl get hpa
```
O alvo é o `Deployment/goserver-test` com CPU alvo em 25%, escalando entre 1 e 3 réplicas.

### Ingress e TLS (opcional)
- `k8s/ingress.yaml` define um host de exemplo `ingress.example.com` com anotação para NGINX e cert-manager.
- `k8s/cluster-issuer.yaml` cria um `ClusterIssuer` do Let’s Encrypt (ajuste `email`).

Passos:
```bash
# com NGINX Ingress e cert-manager já instalados
kubectl apply -f k8s/cluster-issuer.yaml
kubectl apply -f k8s/ingress.yaml
```
Ajuste o DNS do seu domínio para o IP do Ingress Controller. Em ambiente local (Kind), use um /etc/hosts para apontar `ingress.example.com` ao IP do Ingress.

## Manifests adicionais de estudo
- `k8s/pod.yaml` e `k8s/replicaset.yaml`: exemplos básicos.
- `k8s/statefulset.yaml`: exemplo de StatefulSet com MySQL e `volumeClaimTemplates`.
- `k8s/namespaces/*`: exemplo de RBAC por namespace `dev` (ServiceAccount, Role, RoleBinding, Deployment).

## Desenvolvimento
- `go.mod` define o módulo `hello-go` e Go 1.22.
- `Dockerfile` compila o binário `server` e define o `CMD` para executá-lo.

## Limpeza
```bash
kubectl delete -f k8s/hpa.yaml --ignore-not-found
kubectl delete -f k8s/ingress.yaml --ignore-not-found
kubectl delete -f k8s/cluster-issuer.yaml --ignore-not-found
kubectl delete -f k8s/service.yaml --ignore-not-found
kubectl delete -f k8s/deployment.yaml --ignore-not-found
kubectl delete -f k8s/pvc.yaml --ignore-not-found
kubectl delete -f k8s/secret.yaml --ignore-not-found
kubectl delete -f k8s/configmap-family.yaml --ignore-not-found
kubectl delete -f k8s/configmap-env.yaml --ignore-not-found

kind delete cluster --name estudo-k8s || true
```

## Troubleshooting
- `/healthz` retorna 500 por ~10s após o start. As probes consideram esse comportamento.
- Se o Service for `ClusterIP`, use `kubectl port-forward` para testar localmente.
- Em Kind, `LoadBalancer` não provisiona IP público; prefira Ingress ou `port-forward`.
- Ingress do exemplo usa campos antigos em algumas versões do Kubernetes; ajuste para o formato `service.name`/`service.port.number` caso necessário.
