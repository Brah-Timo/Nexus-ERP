
# Nexus ERP




![Nexus ERP Demo](asset/Nexus-ERP.gif)




```markdown


**Ultra-Professional Algerian ERP System — Single Binary Deployment**

> A complete Enterprise Resource Planning solution built with Go + Vue 3, packaged as a single executable with zero external dependencies (except PostgreSQL).

---

## 🏢 Features

| Module | Capabilities |
|--------|-------------|
| **Accounting (SCF)** | Chart of Accounts (Algerian SCF), Journal Entries, Fixed Assets (linear/diminishing), Bank Reconciliation, Trial Balance, Balance Sheet, Income Statement, Budgets, Cost Centers |
| **HR & Payroll** | Employees, Departments, Attendance, Leave Management, Payroll with **Algerian IRG** brackets, CNAS contributions (9%/26%), G29 export |
| **Sales & CRM** | Leads, Opportunities (Kanban), Quotations, Sales Orders, Invoices (TVA/stamp tax), Customer Aging |
| **Purchase** | Suppliers, RFQs, Purchase Orders (approval workflow), Goods Receipts, **3-Way Matching**, Supplier Evaluations |
| **Inventory** | Items (CMUP/FIFO), Categories, Warehouses, Stock Levels, Movements, Inventory Counts, Valuation |
| **Manufacturing** | Bill of Materials, Work Centers, Manufacturing Orders (MRP), Auto Purchase Suggestions |
| **Projects** | Projects, Tasks, Timesheets, Budget vs Actual tracking |
| **Treasury** | Cash Accounts, Bank Accounts, Cheques (deposit/bounce), Payments, Cash Position |
| **Tax (G50/G29)** | TVA declarations, VAT Register, Monthly G50 generation |
| **Workflow** | Rule-based approval chains across all modules |
| **Reports/BI** | Financial ratios, KPI dashboard, comparative analysis |
| **Settings** | Multi-company, Multi-branch, Multi-currency, Fiscal Years, Numbering, Audit Log |

---

## 🚀 Quick Start

### Prerequisites
- **PostgreSQL 14+** (required — the only external dependency)
```
<img width="1894" height="808" alt="image" src="https://github.com/user-attachments/assets/1ff16096-cc9c-4b97-bcd8-919a67c09f8b" />

- **Go 1.21+** (for building from source)
```bash
Download Go language MSI 
https://golang.org/dl/go1.15.1.windows-amd64.msi
```
<img width="495" height="387" alt="image" src="https://github.com/user-attachments/assets/86578a00-807f-49ef-970a-28c01e74a99e" />

- **Node.js 20+** (for building from source)

<img width="757" height="579" alt="image" src="https://github.com/user-attachments/assets/2bb4c7d7-a4ff-4d86-acf4-07deeaaef8f0" />

```


### Option 1 — Run the Pre-built Binary

```bash
# 1. Extract the zip
unzip nexus-erp-v1.0.0.zip
cd nexus-erp-v1.0.0
```
<img width="370" height="29" alt="image" src="https://github.com/user-attachments/assets/47a51a62-b4bd-465a-b31c-7059a0551771" />

```bash
# 2. Set up environment
cp .env.example .env
# Edit .env: set DATABASE_URL to your PostgreSQL connection string

# 3. Start PostgreSQL (if not running)
# macOS: brew services start postgresql
# Ubuntu: sudo systemctl start postgresql
# Windows: Start PostgreSQL service from Services

# 4. Create the database
psql -U postgres -c "CREATE DATABASE nexus_erp;"
psql -U postgres -c "CREATE USER nexus WITH PASSWORD 'nexus_password';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE nexus_erp TO nexus;"

# 5. Run Nexus ERP (migrations run automatically on first start)
./nexus-erp          # Linux/macOS
nexus-erp.exe        # Windows
```
<img width="1465" height="468" alt="image" src="https://github.com/user-attachments/assets/67e363ab-51bc-404e-8c15-8fb0480bdbe3" />

```bash
# 6. Open in browser
# http://localhost:8080
# Login: admin / Admin@123456
```


### CRM
<img width="1508" height="822" alt="image" src="https://github.com/user-attachments/assets/b9a3bcbc-e46f-43dd-93ac-1cd7ab72c9ef" />

### HR
<img width="1512" height="824" alt="image" src="https://github.com/user-attachments/assets/cae5a3af-1a97-4a31-8aa8-ac08da7cb3d7" />

### Option 2 — Docker Compose (Recommended for production)

```bash
# Clone or extract project
git clone <repo-url> nexus-erp
cd nexus-erp

# Start PostgreSQL + App
docker compose up -d

# View logs
docker compose logs -f nexus-erp

# Open: http://localhost:8080
# Login: admin / Admin@123456

# (Optional) Start with pgAdmin database manager
docker compose --profile tools up -d
# pgAdmin: http://localhost:5050 (admin@nexus-erp.local / pgadmin_password)
```

### Option 3 — Build from Source

```bash
# Clone project
git clone <repo-url> nexus-erp
cd nexus-erp

# Build Linux binary
./scripts/build.sh

# Build Windows .exe
./scripts/build.sh --windows

# Build both + package zip
./scripts/build.sh --all --zip

# Run
./nexus-erp
```

---

## 🔑 Default Credentials

| Field | Value |
|-------|-------|
| URL | http://localhost:8080 |
| Username | `admin` |
| Password | `Admin@123456` |

> **Security**: Change the admin password immediately after first login in Settings → Users.

---

## ⚙️ Configuration

All configuration is via environment variables (`.env` file):

```env
# Required
DATABASE_URL=postgres://nexus:password@localhost:5432/nexus_erp?sslmode=disable
JWT_SECRET=your-secret-key-minimum-32-characters

# Optional
APP_ENV=production     # development | production
APP_PORT=8080
LOG_LEVEL=info
```

See `.env.example` for the complete list.

---

## 🏗️ Architecture

```
nexus-erp/
├── main.go                          # Entry point + go:embed + all API routes
├── go.mod / go.sum                  # Go dependencies
├── Dockerfile                       # Multi-stage Docker build
├── docker-compose.yml               # PostgreSQL + App orchestration
├── .env.example                     # Environment variables template
├── migrations/
│   └── 0001_init_schema.sql         # Full PostgreSQL schema (auto-applied)
├── internal/
│   ├── database/db.go               # pgxpool + migration runner
│   ├── middleware/auth_middleware.go # JWT validation
│   ├── models/models.go             # All 50+ ERP data models
│   └── handler/
│       ├── handler.go               # Handler registry + Auth + Dashboard + Settings
│       ├── accounting.go            # Accounting module (COA, JE, Assets, Reports)
│       ├── sales.go                 # Sales/CRM (Customers, Invoices, Pipeline)
│       ├── purchase_inventory.go    # Purchase + Inventory
│       └── hr_mfg_projects_treasury.go  # HR, Manufacturing, Projects, Treasury, Tax
├── scripts/
│   └── build.sh                     # Automated build script
└── web/                             # Vue 3 frontend (embedded at build time)
    ├── index.html
    ├── package.json
    ├── vite.config.ts
    ├── tailwind.config.js
    ├── tsconfig.json
    └── src/
        ├── main.ts                  # App entry
        ├── App.vue                  # Root component
        ├── router/index.ts          # 50+ routes with auth guard
        ├── api/client.ts            # Axios + typed API namespaces
        ├── stores/
        │   ├── auth.ts              # JWT auth store
        │   └── app.ts               # UI state store
        ├── components/
        │   ├── layout/              # AppLayout, Sidebar, AppBar, StatusBar
        │   ├── ui/                  # KpiCard, DataTable, Modal
        │   └── CommandPalette.vue   # Ctrl+K search
        └── modules/                 # Feature modules
            ├── auth/                # Login, ForgotPassword
            ├── dashboard/           # KPI Dashboard
            ├── accounting/          # COA, JE, Reports, Assets, Budgets
            ├── sales/               # Invoices, Customers, Pipeline
            ├── purchase/            # PO, Suppliers, GRN
            ├── inventory/           # Items, Stock, Warehouses
            ├── hr/                  # Employees, Payroll, Leave
            ├── manufacturing/       # BOM, MO, MRP
            ├── projects/            # Projects, Tasks, Timesheets
            ├── treasury/            # Cash, Bank, Cheques, Payments
            ├── tax/                 # G50, VAT Register
            ├── settings/            # Companies, Users, Config
            └── reports/             # Financial Reports
```

### Key Architecture Decisions

1. **Single Binary**: `go:embed web/dist` bundles the entire Vue SPA into the Go binary. No separate web server needed.
2. **Zero ORM**: Raw SQL via `pgx/v5` for maximum performance and control over complex ERP queries.
3. **JWT RS256**: Stateless authentication with access token (24h) + refresh token (30d).
4. **Transactional Integrity**: Critical operations (invoice confirm, payroll, stock movements) run in single PostgreSQL transactions.
5. **Algerian Compliance**: IRG brackets, CNAS rates, SCF chart of accounts, G50/G29 declarations built-in.

---


Based on the features of Nexus ERP from the provided GitHub repository and other major global ERP solutions, here is a detailed comparison table in English.

### Nexus ERP vs. Global ERP Systems: A Feature Comparison

This table compares **Nexus ERP** (based on its v1.0.0 release) against five leading global ERP solutions: **SAP S/4HANA** , **Oracle Fusion Cloud ERP** , **Microsoft Dynamics 365** , **Odoo** , and **ERPNext** .

| Feature / Aspect | **Nexus ERP** | **SAP S/4HANA** | **Oracle Cloud ERP** | **Microsoft Dynamics 365** | **Odoo** | **ERPNext** |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Target Market & Philosophy** | Algerian businesses seeking a cost-effective, compliant, and easy-to-deploy solution. | Large enterprises needing a comprehensive, integrated, and highly customizable global platform . | Large enterprises focused on cloud-based, AI-driven financial management and a unified data model . | Mid-sized to large organizations looking for a scalable, AI-enhanced suite deeply integrated with Microsoft tools . | Small to medium-sized businesses wanting a modern, modular, and cost-effective open-source solution . | Small to large businesses seeking an affordable, open-source, and highly flexible all-in-one system . |
| **Deployment & Architecture** | **Single Binary** (Go + Vue 3) – Extremely simple. Requires only PostgreSQL. | Complex, multi-tier on-premise or cloud. Requires significant infrastructure and expertise . | Cloud-native (SaaS) or on-premise. Requires substantial infrastructure and specialized knowledge . | Cloud (SaaS) or on-premise. Leverages Azure. Integrates deeply with Microsoft ecosystem . | On-premise, Cloud (SaaS), or custom. Modular architecture built with Python/JavaScript . | On-premise or Cloud (SaaS). Built on the Frappe framework (Python/JS) . |
| **Pricing Model** | **Self-Hosted (Free)** – Open source. Cost primarily for internal IT/PostgreSQL hosting. | **Proprietary & Expensive** – High licensing, implementation, and maintenance costs . | **Proprietary & High-Cost** – Subscription-based, significant investment for enterprise-scale features and support. | **Proprietary (SaaS Subscriptions)** – Per-user/per-month; costs vary significantly by module . | **Open Source (Free)** – Community version free. Enterprise version has paid subscriptions for support and hosting. | **Open Source (Free)** – Self-hosted is free. Paid support and cloud hosting available. |
| **Algerian Compliance** | **Native & Built-in** – Pre-loaded SCF Chart of Accounts, IRG tax brackets, CNAS rates, G50/G29 declaration generation. | **Requires Localization** – Heavy customization, partner implementation, and add-ons needed for local legal and fiscal requirements. | **Requires Localization** – Requires significant customization and local partner expertise for Algerian regulations. | **Requires Localization** – Needs configuration or partner-developed extensions for Algerian law. | **Requires Localization** – Requires community-developed or custom-coded modules for Algerian rules. | **Requires Localization** – Requires custom development to adapt to Algerian SCF, IRG, and CNAS. |
| **Key Modules & Breadth** | 14 core modules (Accounting, HR/Payroll, Sales, Purchase, Inventory, Manufacturing, Projects, Treasury, Tax, Workflow, etc.). | Full Suite (Finance, HR, SCM, Manufacturing, Sales, etc.) but highly modular and customizable . | Full Suite (Financials, Project Management, Procurement, Risk Management, etc.) . | Full Suite (Sales, Customer Service, Finance, Supply Chain, HR, Field Service, etc.) . | **Extensive (70+)** – Extremely broad, with apps for everything from eCommerce to POS to Manufacturing . | Comprehensive (Accounting, HR, Inventory, Sales, Purchase, Projects, etc.) . |
| **Ease of Use / UI** | Modern, Single-Page Application built with Vue 3 + Tailwind CSS (Standard modern UI). | Generally considered complex, requiring significant training; older UI being modernized in S/4HANA. | Modern, cloud-native user experience. | Modern, intuitive interface consistent with Windows and other Microsoft products . | Modern, clean, and highly user-friendly interface, known for good usability . | **Functional but Outdated** – Often noted as having an older, less modern interface . |
| **Technical Architecture** | Cutting-edge: **Golang + Vue 3**. Single binary deployment with embedded frontend. Zero ORM for performance. | Proprietary ABAP/Java; complex. S/4HANA optimized for SAP's HANA in-memory database . | Proprietary Java/Cloud; designed for cloud-native, AI-driven processes . | Microsoft Azure; primarily uses .NET and integrates with Azure AI and Power Platform . | Python/JavaScript on the Odoo framework. Highly modular and customizable . | Python/JavaScript on the Frappe framework, which provides a modern backend and a Vue-based UI . |
| **Pros** | **Algerian-specific compliance**, extremely simple deployment, modern tech stack, cost-effective, good feature coverage for core ERP. | Unmatched global enterprise scale, deep process integration, industry best practices built-in. | Strong AI/ML capabilities, native cloud platform, real-time data accessibility, integrated SaaS. | Seamless integration with Microsoft ecosystem (Outlook, Excel), strong AI (Copilot), and powerful BI tools . | Highly modular, extensive app ecosystem, user-friendly, democratizes access to ERP for SMBs . | Very cost-effective, open-source, highly customizable, comprehensive core features . |
| **Cons** | **New system** with limited global market presence. Lacks niche or industry-specific modules (e.g., Fleet, Quality) mentioned in the release notes. | High cost, long and complex implementation cycles, requires specialized consultants, less flexible . | High cost, complex to implement, can be overkill for smaller businesses. | Can become expensive with multiple modules; implementation often requires a partner for complex scenarios. | Complexity can grow significantly with many modules; enterprise-grade customization may require partner support. | Initial setup and deep customization require **technical knowledge**; interface can feel dated . |

### Summary

- **Nexus ERP** is a highly specialized, modern, and cost-effective solution that is **uniquely positioned for the Algerian market**. Its key strength is its **native compliance** with local legal and fiscal requirements (SCF, IRG, CNAS, G50/G29)  combined with a **simple, modern, and powerful** technical architecture (single binary, Go+Vue).
- **SAP, Oracle, and Microsoft** are "global giants" designed for large enterprises. They offer immense depth and scale but come with very high costs and complexity. They are not tailored for Algeria out-of-the-box and would require expensive customizations .
- **Odoo and ERPNext** are direct open-source competitors. They offer a similar cost benefit and modularity. However, Nexus ERP is arguably more modern in its architecture (single binary vs. traditional application stacks) and has a specific, built-in focus on the **Algerian market** that Odoo and ERPNext lack without significant custom development .

In conclusion, **Nexus ERP competes by offering a "best of both worlds" proposition**: the affordability and openness of solutions like Odoo, with a state-of-the-art, simple-to-deploy architecture and immediate, out-of-the-box compliance for **Algerian businesses** .

---

## 🇩🇿 Algerian Compliance

| Regulation | Implementation |
|-----------|----------------|
| **SCF** (Plan Comptable Financier) | Full 7-class chart of accounts pre-loaded |
| **IRG** (Impôt sur le Revenu Global) | 6-bracket progressive tax calculation |
| **CNAS** | Employee 9% + Employer 26% contributions |
| **TVA** | 19% standard rate, deductible/collected tracking |
| **G50** | Monthly TVA declaration generation |
| **G29** | Monthly CNAS/IRG payroll declaration export |
| **TAP** | Taxe sur l'Activité Professionnelle support |
| **NIF/NIS/RC/ART** | Company registration fields |

---

## 🔧 API Reference

All API endpoints are under `/api/` prefix. Authentication via `Authorization: Bearer <token>`.

### Auth
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/login` | Login (returns access_token + refresh_token) |
| POST | `/api/auth/logout` | Logout |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/forgot-password` | Request password reset |

### Core Modules
| Module | Base Path | Key Endpoints |
|--------|-----------|---------------|
| Dashboard | `/api/dashboard` | `/summary`, `/cashflow`, `/activity` |
| Accounting | `/api/accounting` | `/chart-of-accounts`, `/journal-entries`, `/trial-balance`, `/balance-sheet`, `/income-statement` |
| Sales | `/api/sales` | `/customers`, `/invoices`, `/quotations`, `/orders`, `/pipeline` |
| Purchase | `/api/purchase` | `/suppliers`, `/purchase-orders`, `/goods-receipts`, `/invoices` |
| Inventory | `/api/inventory` | `/items`, `/warehouses`, `/stock-levels`, `/movements` |
| HR | `/api/hr` | `/employees`, `/departments`, `/leave-requests`, `/payroll` |
| Manufacturing | `/api/manufacturing` | `/bom`, `/manufacturing-orders`, `/mrp-suggestions` |
| Projects | `/api/projects` | `/projects`, `/tasks`, `/timesheets` |
| Treasury | `/api/treasury` | `/cash-accounts`, `/bank-accounts`, `/cheques`, `/payments` |
| Tax | `/api/tax` | `/g50`, `/vat-register` |
| Reports | `/api/reports` | `/financial-ratios`, `/aging` |
| Workflow | `/api/workflow` | `/rules`, `/approvals`, `/approve`, `/reject` |
| Settings | `/api/settings` | `/companies`, `/users`, `/fiscal-years`, `/currencies` |

---

## 🏭 Production Deployment

### Linux Server

```bash
# 1. Build binary
./scripts/build.sh

# 2. Create system user
sudo useradd -r -s /bin/false nexus-erp

# 3. Copy binary
sudo cp nexus-erp /usr/local/bin/
sudo chmod +x /usr/local/bin/nexus-erp

# 4. Create config
sudo mkdir -p /etc/nexus-erp
sudo cp .env.example /etc/nexus-erp/.env
sudo nano /etc/nexus-erp/.env  # Edit settings

# 5. Create systemd service
sudo tee /etc/systemd/system/nexus-erp.service > /dev/null <<EOF
[Unit]
Description=Nexus ERP
After=network.target postgresql.service

[Service]
User=nexus-erp
EnvironmentFile=/etc/nexus-erp/.env
ExecStart=/usr/local/bin/nexus-erp
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 6. Enable and start
sudo systemctl daemon-reload
sudo systemctl enable nexus-erp
sudo systemctl start nexus-erp
sudo systemctl status nexus-erp
```

### Windows Server

```powershell
# 1. Build .exe
./scripts/build.sh --windows

# 2. Copy nexus-erp.exe to C:\NexusERP\
# 3. Copy .env.example to C:\NexusERP\.env and configure
# 4. Register as Windows Service (using NSSM)
nssm install NexusERP "C:\NexusERP\nexus-erp.exe"
nssm set NexusERP AppDirectory "C:\NexusERP"
nssm start NexusERP
```

### Nginx Reverse Proxy (optional)

```nginx
server {
    listen 80;
    server_name erp.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name erp.yourdomain.com;

    ssl_certificate     /etc/letsencrypt/live/erp.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/erp.yourdomain.com/privkey.pem;

    client_max_body_size 50M;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
    }
}
```

---

## 🛠️ Development

```bash
# Terminal 1: Go backend (with hot reload using air)
cd nexus-erp
go install github.com/air-verse/air@latest
air

# Terminal 2: Vue frontend (Vite dev server with HMR)
cd nexus-erp/web
npm install
npm run dev
# → http://localhost:5173 (proxies /api to :8080)

# Type check
cd web && npm run type-check

# Lint
cd web && npm run lint
```

---

## 📋 Changelog

### v1.0.0 (2026-08-08)
- Initial release
- Full Algerian SCF/IRG/CNAS compliance
- 14 modules: Accounting, HR/Payroll, Sales/CRM, Purchase, Inventory, Manufacturing, Projects, Treasury, Tax, Workflow, Reports, Dashboard, Settings
- Single binary deployment with go:embed
- 100+ REST API endpoints
- Vue 3 + TypeScript + Tailwind CSS frontend
- Command palette (Ctrl+K), real-time approval badges

---

## 📄 License

Copyright © 2026 Nexus ERP. All rights reserved.

---

> Built with ❤️ for Algerian businesses.
```
