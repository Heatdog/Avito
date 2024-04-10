# insert examples

# ------------------------ OK ------------------------

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "some_title", "text": "some_text", "url": "some_url"},
  "feature_id": 1,
  "is_active": true,
  "tag_id": [
    1,2,3
  ]
}'

# response {"banner_id":1}

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "some_title", "text": "some_text", "url": "some_url"},
  "feature_id": 2,
  "is_active": true,
  "tag_id": [
    2
  ]
}'

# response {"banner_id":2}

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "good_title"},
  "feature_id": 2,
  "is_active": true,
  "tag_id": [
    3,4,5
  ]
}'

# response {"banner_id":3}

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "good_title"},
  "feature_id": 3,
  "is_active": true,
  "tag_id": [
    4
  ]
}'

# response {"banner_id":3}

# ------------------------ Unauthorized ------------------------

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "good_title"},
  "feature_id": 3,
  "is_active": true,
  "tag_id": [
    4
  ]
}'

# response 401 Error: Unauthorized

# ------------------------ Forbidden ------------------------

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: user_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "good_title"},
  "feature_id": 3,
  "is_active": true,
  "tag_id": [
    4
  ]
}'

# response 403 Error: Forbidden

# ------------------------ validation error ------------------------
 
curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "feature_id": 3,
  "is_active": true,
  "tag_id": [
    4
  ]
}'

# response {"error":"Key: 'BannerInsert.Content' Error:Field validation for 'Content' failed on the 'json' tag"}

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": 123,
  "feature_id": 3,
  "is_active": true,
  "tag_id": [
    4
  ]
}'

# response {"error":"Key: 'BannerInsert.Content' Error:Field validation for 'Content' failed on the 'json' tag"}

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "good_title"},
  "is_active": true,
  "tag_id": [
    4
  ]
}'

# response {"error":"Key: 'BannerInsert.FeatureID' Error:Field validation for 'FeatureID' failed on the 'required' tag"}

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "good_title"},
  "feature_id": 3,
  "tag_id": [
    4
  ]
}'

# response {"error":"Key: 'BannerInsert.IsActive' Error:Field validation for 'IsActive' failed on the 'required' tag"}

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "good_title"},
  "feature_id": 3,
  "is_active": true
}'

# response {"error":"Key: 'BannerInsert.TagsID' Error:Field validation for 'TagsID' failed on the 'required' tag"}

# ------------------------ dublicate error ------------------------

curl -X 'POST' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token' \
  -H 'Content-Type: application/json' \
  -d '{
  "content": {"title": "good_title"},
  "feature_id": 3,
  "is_active": true,
  "tag_id": [
    4
  ]
}'

# response {"error":"ERROR: duplicate key value violates unique constraint \"features_tags_to_banners_pk\" (SQLSTATE 23505)"}