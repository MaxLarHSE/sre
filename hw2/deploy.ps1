$ErrorActionPreference = 'Stop'
$k8s = Join-Path $PSScriptRoot 'k8s'

function Step([string]$cmd) {
    Write-Host "`n> $cmd" -ForegroundColor Cyan
    Invoke-Expression $cmd
    if ($LASTEXITCODE -ne 0) { throw "failed: $cmd" }
}

Step "kubectl apply -f $k8s/00-namespace.yml -f $k8s/01-secret.yml -f $k8s/02-configmap.yml"

Step "kubectl apply -f $k8s/10-postgres-statefulset.yml -f $k8s/11-postgres-service.yml"
Step "kubectl -n trade rollout status statefulset/postgres --timeout=180s"

Step "kubectl -n trade delete job trade-migrate --ignore-not-found"
Step "kubectl apply -f $k8s/20-migrate-job.yml"
Step "kubectl -n trade wait --for=condition=complete job/trade-migrate --timeout=180s"

Step "kubectl apply -f $k8s/30-app-deployment.yml -f $k8s/31-app-service.yml"
Step "kubectl -n trade rollout status deployment/trade-app --timeout=180s"

Step "kubectl -n trade get all"
