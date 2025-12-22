# Nexus

Verified professional identity powered by DIDs and verifiable credentials.

## Structure

```
nexus/
├── platform/   # Go API, workers, domain logic
├── frontend/   # Next.js web application
└── infra/      # Docker, K8s, Terraform
```

## Getting Started

### Platform (Go)

```bash
cd platform
make run
```

### Frontend (Next.js)

```bash
cd frontend
npm install
npm run dev
```

## License

MIT
