# Admin Panel — Login & Users

The admin panel is a small web app at `http://<your-host>:3104` (port `3104` by default). All settings — projects, payment methods, routes — are managed here.

## First login

When the stack starts for the first time, control-api creates a single **superadmin** account using the `SUPERADMIN_USERNAME` and `SUPERADMIN_PASSWORD` environment variables. In a fresh `docker-compose.yml` both are set to `superadmin`.

Open the login page, enter the credentials, and submit.

> **Screenshot placeholder:** ![Login page](./../images/login.png)

After login you land on the Dashboard. The left sidebar gives you access to every other section.

> **Screenshot placeholder:** ![Sidebar menu](./../images/sidebar-menu.png)

> **Important.** Change the default password before exposing the panel to the internet. You can do it in two ways:
> - Edit `SUPERADMIN_PASSWORD` in `docker-compose.yml` and restart `control-api`. This works only for the bootstrap account.
> - Or log in, open **Users**, edit the superadmin record and set a new password from the form.

## Roles

There are two roles:

| Role | What they can do |
|---|---|
| **Superadmin** | Everything. Manage other users, all projects, all routes, all payment configuration. |
| **User** | Manage their own projects and routes. Cannot create or edit other users. Cannot edit projects they do not own. |

The **Users** section is visible only to superadmins.

## Creating a user

Only superadmins can create users. Open **Users** in the sidebar and click **Create User**.

> **Screenshot placeholder:** ![Users list](./../images/users-list.png)

Fill in:

- **Username** — what the user types into the login form. Must be unique.
- **Password** — set an initial password. The user can change it later from their own profile.
- **Role** — `superadmin` or `user`.

> **Screenshot placeholder:** ![User form](./../images/users-form.png)

Save. The user can now log in with the credentials you set.

## Editing and deleting users

From the Users list, the row actions let you view, edit or delete each account. Deleting a user does **not** delete the projects they owned — the projects remain and become editable only by superadmins, until you reassign them.

> **Caution.** You cannot delete the only remaining superadmin. Always keep at least one. If you lose access to it, you have to restore from a database backup or set `SUPERADMIN_USERNAME` / `SUPERADMIN_PASSWORD` in `docker-compose.yml` and restart control-api — the bootstrap logic recreates the account on next boot if it is missing.

## What's next?

- See what the home page shows: [Dashboard](./02-dashboard.md).
- Start setting up payments: [Facilitators](./03-facilitators.md).
