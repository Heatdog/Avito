# get examples

# ------------------------ OK ------------------------

curl -X 'GET' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

#response 200
:`
[{"banner_id":9,"tag_ids":[1,2,3],"feature_id":4,"content_v1":{"title":"good_title3"},"content_v2":null,"content_v3":null,
"is_active":false,"created_at":"2024-04-10T14:53:04.260146Z","updated_at":"2024-04-10T14:53:04.260146Z"},
{"banner_id":7,"tag_ids":[1,2,3],"feature_id":5,"content_v1":{"title":"good_title2"},
"content_v2":null,"content_v3":null,"is_active":true,"created_at":"2024-04-10T14:51:38.587115Z",
"updated_at":"2024-04-10T14:51:38.587115Z"},{"banner_id":4,"tag_ids":[4],"feature_id":3,
"content_v1":{"title":"good_title"},"content_v2":null,"content_v3":null,"is_active":true,
"created_at":"2024-04-10T14:26:17.217056Z","updated_at":"2024-04-10T14:26:17.217056Z"},
{"banner_id":3,"tag_ids":[3,4,5],"feature_id":2,"content_v1":{"title":"good_title"},
"content_v2":null,"content_v3":null,"is_active":true,"created_at":"2024-04-10T14:25:19.889336Z",
"updated_at":"2024-04-10T14:25:19.889336Z"},{"banner_id":2,"tag_ids":[2],"feature_id":2,
"content_v1":{"text":"some_text","title":"some_title","url":"some_url"},"content_v2":null,
"content_v3":null,"is_active":true,"created_at":"2024-04-10T14:24:15.121371Z","updated_at":"2024-04-10T14:24:15.121371Z"},
{"banner_id":1,"tag_ids":[1,2,3],"feature_id":1,"content_v1":{"text":"some_text","title":"some_title","url":"some_url"},
"content_v2":null,"content_v3":null,"is_active":true,"created_at":"2024-04-10T14:22:46.636712Z",
"updated_at":"2024-04-10T14:22:46.636712Z"}]
`

# only tag

curl -X 'GET' \
  'http://localhost:8080/banner?tag_id=1' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

#response 200
:`
[{"banner_id":9,"tag_ids":[1,2,3],"feature_id":4,"content_v1":{"title":"good_title3"},"content_v2":null,
"content_v3":null,"is_active":false,"created_at":"2024-04-10T14:53:04.260146Z","updated_at":"2024-04-10T14:53:04.260146Z"},
{"banner_id":7,"tag_ids":[1,2,3],"feature_id":5,"content_v1":{"title":"good_title2"},"content_v2":null,
"content_v3":null,"is_active":true,"created_at":"2024-04-10T14:51:38.587115Z","updated_at":"2024-04-10T14:51:38.587115Z"},
{"banner_id":1,"tag_ids":[1,2,3],"feature_id":1,"content_v1":{"text":"some_text","title":"some_title","url":"some_url"},
"content_v2":null,"content_v3":null,"is_active":true,"created_at":"2024-04-10T14:22:46.636712Z",
"updated_at":"2024-04-10T14:22:46.636712Z"}]
`

# tag and feature

curl -X 'GET' \
  'http://localhost:8080/banner?tag_id=1&feature_id=1' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

#response 200
:`
[{"banner_id":1,"tag_ids":[1,2,3],"feature_id":1,"content_v1":{"text":"some_text","title":"some_title","url":"some_url"},
"content_v2":null,"content_v3":null,"is_active":true,"created_at":"2024-04-10T14:22:46.636712Z",
"updated_at":"2024-04-10T14:22:46.636712Z"}]
`

# with limit

curl -X 'GET' \
  'http://localhost:8080/banner?limit=2' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

#response 200
:`
[{"banner_id":9,"tag_ids":[1,2,3],"feature_id":4,"content_v1":{"title":"good_title3"},"content_v2":null,
"content_v3":null,"is_active":false,"created_at":"2024-04-10T14:53:04.260146Z","updated_at":"2024-04-10T14:53:04.260146Z"},
{"banner_id":7,"tag_ids":[1,2,3],"feature_id":5,"content_v1":{"title":"good_title2"},"content_v2":null,"content_v3":null,
"is_active":true,"created_at":"2024-04-10T14:51:38.587115Z","updated_at":"2024-04-10T14:51:38.587115Z"}]
`
# with limit and offset

curl -X 'GET' \
  'http://localhost:8080/banner?limit=1&offset=1' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

#response 200
:`
[{"banner_id":7,"tag_ids":[1,2,3],"feature_id":5,"content_v1":{"title":"good_title2"},"content_v2":null,"content_v3":null,
"is_active":true,"created_at":"2024-04-10T14:51:38.587115Z","updated_at":"2024-04-10T14:51:38.587115Z"}]
`

# no banners

curl -X 'GET' \
  'http://localhost:8080/banner?tag_id=2132131' \
  -H 'accept: application/json' \
  -H 'token: admin_token'

#response 200
:`
null
`

# ------------------------ Unauthorized ------------------------

curl -X 'GET' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json'

# response 401 Error: Unauthorized

# ------------------------ Forbidden ------------------------

curl -X 'GET' \
  'http://localhost:8080/banner' \
  -H 'accept: application/json' \
  -H 'token: user_token'

# response 403	Error: Forbidden

