# Data Model: Enhanced Student Profile

## New Type: Gender

| Value | Display |
|-------|---------|
| `male` | Male |
| `female` | Female |
| `other` | Prefer not to say |

Nullable — existing users default to `NULL`.

## Modified Entity: User

```
users
├── id              uuid PK
├── name            varchar
├── email           varchar UNIQUE
├── password_hash   varchar
├── role            user_role enum
├── status          user_status enum
├── education_level education_level enum (nullable)
├── gender          gender enum ← NEW (nullable)
├── phone           varchar(20)  ← NEW (nullable)
├── avatar_url      text         ← NEW (nullable)
├── created_at      timestamptz
└── updated_at      timestamptz
```

**Validation rules**:
- `gender`: must be one of `male`, `female`, `other`, or null
- `phone`: digits only (strip non-digits server-side), 10-15 digits
- `avatar_url`: no server-side validation (set from trusted upload endpoint)
- All three fields are nullable — existing users have no values

## Migration

**File**: `000022_profile_fields.up.sql`

```sql
CREATE TYPE gender AS ENUM ('male', 'female', 'other');
ALTER TABLE users ADD COLUMN gender gender;
ALTER TABLE users ADD COLUMN phone varchar(20);
ALTER TABLE users ADD COLUMN avatar_url text;
```

**Down**: `000022_profile_fields.down.sql`

```sql
ALTER TABLE users DROP COLUMN IF EXISTS gender;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
DROP TYPE IF EXISTS gender;
```
