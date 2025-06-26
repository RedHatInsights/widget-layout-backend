# Developer Guide: Making Requests Locally with Identity Header

When developing or testing the Widget Layout Backend locally, most endpoints require a valid `x-rh-identity` header. This guide explains how to generate and use this header for local requests.

## 1. Generating the Identity Header

A helper script is provided at `cmd/dev/user-identity.go` to generate a valid identity header for local development. You can use the provided Makefile target to generate it easily.

### Steps:

1. **Generate the Identity Header Using Makefile**

   ```sh
   make generate-identity
   ```

   This will output a long base64-encoded string. This is your identity header value.

2. **Copy the Output**

   Save the output string for use in your API requests.

## 2. Making Requests with the Identity Header

When using `curl`, Postman, or any HTTP client, include the header as follows:

### Example with `curl`:

```sh
curl -H "x-rh-identity: <PASTE_IDENTITY_HEADER_HERE>" http://localhost:PORT/api/widget-layout/v1/widgets
```

Replace `<PASTE_IDENTITY_HEADER_HERE>` with the string generated in step 1, and `PORT` with the port your server is running on (default is usually 8000 or as configured).

### Example with Postman:
- In the Headers tab, add:
  - Key: `x-rh-identity`
  - Value: *(paste the generated string)*

## 3. Regenerating the Header

If you need a different user or want to reset the identity, you can edit the values in `cmd/dev/user-identity.go` (e.g., `UserID`, `AccountNumber`) and rerun the script using:

```sh
make generate-identity
```

## 4. Troubleshooting
- If you receive a `400 Bad Request` with "Invalid identity header", ensure the header is present and correctly generated.
- The backend will reject requests without a valid `x-rh-identity` header.

## 5. Example Output

A valid header will look like:

```
eyJpZGVudGl0eSI6eyJ1c2VyIjp7IlVzZXJJRCI6InVzZXItMTIzIiwiRmlyc3ROYW1lIjoiSm9obiIsIkxhc3ROYW1lIjoiRG9lIn0sImFjY291bnRfbnVtYmVyIjoiMTIzNDU2Nzg5MCJ9LCJlbnRpdGxlbWVudHMiOnt9fQ==
```

---

**Summary:**
- Use `make generate-identity` to generate the header.
- Add the output as the `x-rh-identity` header in your requests.
- All local API requests must include this header.
