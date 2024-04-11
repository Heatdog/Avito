# update example

# all fields update

curl -X 'PATCH' \
  'http://localhost:8080/banner/4' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"test":"test"},
  "feature_id": 1,
  "is_active": true,
  "tag_id": [
    2
  ]
}'

# 200


# active field update only

curl -X 'PATCH' \
  'http://localhost:8080/banner/4' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "is_active": true
}'

# 200

# tag field update only

curl -X 'PATCH' \
  'http://localhost:8080/banner/4' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "tag_id": [
    1,2
  ]
}'

# 200

# feature_id update only

curl -X 'PATCH' \
  'http://localhost:8080/banner/4' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "feature_id": 3
}'

# dublicated keys error

curl -X 'PATCH' \
  'http://localhost:8080/banner/4' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "feature_id": 1
}'

# 500 {"error":"ERROR: duplicate key value violates unique constraint \"features_tags_to_banners_pk\" (SQLSTATE 23505)"}

# banner not found

curl -X 'PATCH' \
  'http://localhost:8080/banner/1312321312' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "feature_id": 1
}'

# 500 {"error":"ERROR: duplicate key value violates unique constraint \"features_tags_to_banners_pk\" (SQLSTATE 23505)"}

# ------------------- set last version --------------------

curl -X 'PATCH' \
  'http://localhost:8080/banner/4/2' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

# 200

# bad version

curl -X 'PATCH' \
  'http://localhost:8080/banner/4/231' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

# 400 {"error":"bad version"}

# not found

curl -X 'PATCH' \
  'http://localhost:8080/banner/4/2' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

# 404