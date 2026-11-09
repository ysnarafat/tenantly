# Tenantly Operations Portal

The Operations Portal is a separately built and deployed Angular application for Tenantly platform staff. It is intentionally separate from the property-management app and is not included in the public marketing-site routes.

## What it manages

- Three bilingual public plans (English and Bangla names, descriptions, feature lists), BDT pricing, limits, visibility, and display order.
- Organization/customer contact lookup and manual subscription tier, status, and renewal-date changes.
- An internal organization-linked support queue with status, staff assignment, customer-visible replies, and separate staff-only notes.

Organization admins use the Support Center in the main Tenantly application to submit tickets and reply to the support team. Tenant-facing support routes require an active `ORG_ADMIN` membership in the organization selected by the authenticated session; the `/api/v1/ops` staff routes remain `SUPER_ADMIN`-only.

Prices are entered in taka and stored as integer paisa. A blank price displays a contact CTA on the marketing site. The initial limits are conservative metadata defaults and should be configured by staff before launch; product-side quota enforcement is not included yet. Online checkout and invoice/payment processing are outside this first release.

## Organization onboarding

Platform staff can create an organization from the portal, choose its initial plan, and invite the first `ORG_ADMIN`. Organization and invitation records are created together; the organization subscription is initialized by the database trigger. The invited admin receives a single-use link, enters their name, and chooses their own password before the account is created and added to the organization. Invitation links expire after seven days.

The API sends invitations through SMTP using `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, and `FROM_EMAIL`. SMTP must support STARTTLS. Set `INVITATION_APP_URL` to the customer application's origin (for example, `https://app.tenantly.example`); the API appends `/accept-invitation`. Delivery failures are recorded in the portal so staff can resend, which rotates the token and invalidates the previous link. Pending invitations can also be revoked. Configure real SMTP credentials and the production app URL before enabling onboarding in production.

## Local development

From `src/frontend`:

```bash
npm ci
npm run start:operations
```

Open `http://localhost:4300`. The development API URL defaults to `http://localhost:8080/api/v1`. The backend's development CORS defaults include ports 4200 and 4300. Sign in with an account whose platform role is `SUPER_ADMIN`; API authorization is enforced independently of the portal UI.

Build the standalone portal with:

```bash
npm run build:operations
```

The output is `dist/operations-portal`. `npm run build` continues to build only the customer-facing application.
Run staff portal unit tests with `npm run test:operations`.

## Deployment and security

Deploy this output as its own static site and hostname (for example, `ops.tenantly.example`), with SPA fallback to `index.html`. The production API URL is `/api/v1`; configure the reverse proxy to route that path to the API, or change `operations-portal/src/environments/environment.prod.ts` to the API origin.

Set the API's `CORS_ALLOWED_ORIGINS` to the exact operations hostname and the marketing/customer-app origins. Do not use a wildcard in production. Keep the portal hostname access-controlled at the edge where practical, and provision platform-administrator accounts only for staff.

The browser stores the short-lived access token in session storage and clears it on sign-out. Backend routes under `/api/v1/ops` require `SUPER_ADMIN`; hiding a portal control is never the authorization boundary. Public pricing is served separately by the read-only `/api/v1/public/plans` endpoint.
