# update example

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

curl -X 'PATCH' \
  'http://localhost:8080/banner/4/2' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

# 200