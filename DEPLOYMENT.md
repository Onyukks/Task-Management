# Deployment runbook

Architecture: **Neon** (Postgres) ← **Render** (Go API, Docker) ← **Vercel**
(Next.js frontend). Do the steps in order — the frontend needs the API URL, and
the API needs the frontend URL for CORS.

> All three services have free tiers. Note: Render free instances sleep after
> ~15 min idle, so the very first request after a nap takes a few seconds to
> wake. Neon free also autosuspends. Fine for a demo; mention it if a reviewer
> hits a cold start.

---

## 0. Push to GitHub

```bash
git remote add origin https://github.com/<you>/<repo>.git
git push -u origin main
```

---

## 1. Database — Neon

1. Create a project at https://neon.tech (any region).
2. Copy the **pooled** connection string (it contains `-pooler`), e.g.
   `postgres://user:pass@ep-xxx-pooler.region.aws.neon.tech/neondb?sslmode=require`.
   Keep it for step 2. Migrations run automatically on the API's first boot.

---

## 2. Backend — Render

1. https://render.com → **New → Blueprint** → connect the repo. Render reads
   [`render.yaml`](render.yaml) and creates the `tasks-api` web service.
2. Set the two `sync:false` env vars:
   - `DATABASE_URL` → the Neon pooled string from step 1.
   - `FRONTEND_ORIGIN` → leave a placeholder for now (e.g. `https://example.com`);
     you'll update it in step 4.
3. Deploy. When live, note the URL, e.g. `https://tasks-api.onrender.com`.
   Verify: `curl https://tasks-api.onrender.com/healthz` → `{"status":"ok"}`.

(No `gh`/blueprint? You can also create a **Web Service** manually: runtime
Docker, root `backend`, health check `/healthz`, and add the env vars by hand.)

---

## 3. Frontend — Vercel

1. https://vercel.com → **Add New → Project** → import the repo.
2. **Root Directory: `frontend`** (important — it's a monorepo).
3. Environment variable: `NEXT_PUBLIC_API_URL` → your Render URL from step 2
   (e.g. `https://tasks-api.onrender.com`, no trailing slash).
4. Deploy. Note the URL, e.g. `https://tasks-app.vercel.app`.

Or from the CLI (already installed here):

```bash
cd frontend
vercel link          # follow prompts
vercel env add NEXT_PUBLIC_API_URL production   # paste the Render URL
vercel --prod
```

---

## 4. Connect them (CORS + cookies)

Back in **Render → tasks-api → Environment**, set `FRONTEND_ORIGIN` to your
Vercel URL (exact, no trailing slash) and let it redeploy. This is required so
the API allows the browser's credentialed requests and sets the cross-site
auth cookie (`Secure` + `SameSite=None`, already enabled by `APP_ENV=production`).

Done. Open the Vercel URL and sign up.

---

## 5. (Optional) Make yourself an admin

In the Neon SQL editor:

```sql
UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
```

Sign out and back in; the "View all tasks" toggle appears in the header.

---

## Updating the README links

After deploying, replace the placeholders at the top of [`README.md`](README.md)
with your real frontend + API URLs before submitting.
