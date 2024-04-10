# delete examples

curl -X 'DELETE' \
  'http://localhost:8080/banner/1' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

# 204

curl -X 'DELETE' \
  'http://localhost:8080/banner/1' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

# 404

curl -X 'DELETE' \
  'http://localhost:8080/banner?tag_id=3&feature_id=2' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

# 202