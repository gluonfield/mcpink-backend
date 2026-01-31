-- +goose Up
-- Convert github_scopes from TEXT to TEXT[] array
ALTER TABLE users
  ALTER COLUMN github_scopes DROP DEFAULT,
  ALTER COLUMN github_scopes TYPE TEXT[] USING
    CASE
      WHEN github_scopes = '' THEN '{}'::TEXT[]
      ELSE string_to_array(regexp_replace(github_scopes, '[,\s]+', ',', 'g'), ',')
    END,
  ALTER COLUMN github_scopes SET DEFAULT '{}';

-- +goose Down
ALTER TABLE users
  ALTER COLUMN github_scopes DROP DEFAULT,
  ALTER COLUMN github_scopes TYPE TEXT USING array_to_string(github_scopes, ' '),
  ALTER COLUMN github_scopes SET DEFAULT '';
