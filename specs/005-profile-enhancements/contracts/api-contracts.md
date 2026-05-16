# API Contracts: Enhanced Student Profile

## Modified Schema: UpdateProfileRequest

Add three new optional fields:

```yaml
UpdateProfileRequest:
  type: object
  properties:
    education_level:
      type: string
      enum: [sd, smp, sma, smk]
      nullable: true
    gender:                          # NEW
      type: string
      enum: [male, female, other]
      nullable: true
    phone:                           # NEW
      type: string
      nullable: true
      description: 10-15 digit phone number
    avatar_url:                      # NEW
      type: string
      nullable: true
      description: URL from POST /api/v1/upload
```

## Modified Schema: UserResponse

Add three new fields:

```yaml
UserResponse:
  # existing fields...
  properties:
    # ...existing...
    gender:                          # NEW
      type: string
      enum: [male, female, other]
      nullable: true
    phone:                           # NEW
      type: string
      nullable: true
    avatar_url:                      # NEW
      type: string
      nullable: true
```

## Endpoint: PATCH /auth/me

No route change. Request body extended with optional `gender`, `phone`, `avatar_url`. Response includes new fields.

## Endpoint: POST /api/v1/upload

No change. Existing endpoint handles image upload (2 MB limit, image types only). The resulting URL is stored via `PATCH /auth/me`.
