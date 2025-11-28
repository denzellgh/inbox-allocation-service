# Inbox Allocation Service

This repository contains the **Inbox Allocation Service**, composed of:

- **backend/**: main service written in Go (Golang)
- **postman-collection/**: Postman collection to test the API

---

## Main backend documentation

The detailed backend documentation (requirements, how to run it locally, environment variables, etc.) is located at:

- `backend/README.md`

Please refer to that file for any technical details about the service.

---

## Postman collection

The `postman-collection/` directory includes:

- `inbox-allocation-api.postman_collection.json`

### How to use the collection

1. Open Postman (desktop or web).
2. Go to **File → Import**.
3. Select the file:
   - `postman-collection/inbox-allocation-api.postman_collection.json`
4. Once the collection is imported:
   - Configure the **environment variable/base URL** to point to the endpoint where the backend is deployed (for example, the public URL of the ECS service).
   - Review and adjust any headers or variables that depend on the environment (for example, `X-Tenant-ID` if applicable).
5. Send the requests in the collection to test the service endpoints.

---

## AWS infrastructure status

The AWS resources (RDS, ECS, ECR, etc.) associated with this project **will remain up for approximately the next 8 hours** from the last deployment.

After that period, **the services will likely be shut down** to avoid unnecessary costs.

If you need the services to be available again at any time (for example, for demos, debugging, or additional tests), **I can bring the infrastructure back up without any problem**.

- Just provide an **email** or **WhatsApp** contact, and we can coordinate to have the infrastructure available when you need it.

---

## Relevant structure

- `/backend` → Service source code (see `backend/README.md`).
- `/postman-collection` → Postman collection to test the API (`inbox-allocation-api.postman_collection.json`).
