# Vercel frontend production deployment

- Owner: Frontend release owner
- Last reviewed: 2026-08-05
- Hosting boundary: Vercel only

Create three Vercel projects from this repository. For each project, select the
listed Root Directory and keep “Include source files outside of the Root
Directory” enabled so pnpm workspace packages remain available.

| Project    | Root Directory   | Required Production variables                      |
| ---------- | ---------------- | -------------------------------------------------- |
| Marketing  | `apps/marketing` | `NEXT_PUBLIC_MARKETING_URL`, `OBIARA_API_BASE_URL` |
| Member web | `apps/web`       | `OBIARA_API_BASE_URL`, `NEXT_PUBLIC_API_BASE_URL`  |
| Admin      | `apps/admin`     | `OBIARA_API_BASE_URL`, `NEXT_PUBLIC_API_BASE_URL`  |

Use Node `22.13.0` and the repository package manager (`pnpm@11.17.0`). Vercel
should detect Next.js and the workspace build automatically. Do not set a
custom output directory. The server-only `OBIARA_API_BASE_URL` should be the
primary API binding; the public duplicate remains for existing client/runtime
compatibility and contains no credential.

The `.env.production.example` file inside each app is safe to commit. Its local
`.env.production` counterpart is ignored and is a worksheet only. Enter the
same values in the Vercel project's **Production** environment; local files are
not a substitute for Vercel configuration.

## Order of operations

1. Create the backend from `render.yaml`, fill all required `sync: false`
   values and manually deploy the verified API/worker commit.
2. Confirm `https://<api-host>/live` and `https://<api-host>/ready`.
3. Replace the API origin in each frontend production worksheet and Vercel
   Production environment.
4. Set `NEXT_PUBLIC_MARKETING_URL` to the final canonical HTTPS domain, not a
   preview deployment URL.
5. Deploy marketing, member web and admin from their respective Root
   Directories.
6. Verify marketing metadata, `/privacy`, `/terms`, `/support`,
   `/delete-account`, waitlist submission, member sign-in and admin sign-in.

Production server routes fail closed when the API origin is absent, preventing
a Vercel deployment from silently calling `127.0.0.1`.
