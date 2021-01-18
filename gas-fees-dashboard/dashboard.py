import os
import requests
import glob
import json
import http.client
from redashAPI import RedashAPIClient
from dotenv import load_dotenv

http.client._MAXLINE = 655360 * 4

load_dotenv()

redash_url = os.getenv('REDASH_URL', 'http://localhost:5000')
setupPayload = {
    'name': os.getenv('USER_NAME', 'admin'), 'email': os.getenv('USE_EMAIL', 'admin@p2p.org'),
    'password': os.getenv('USER_PASS', 'supersecret123'), 'security_notifications': 'y',
    'org_name': os.getenv('ORG_NAME', 'p2p')
}
setupHeaders = {
    'Content-Type': 'application/x-www-form-urlencoded',
    'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,'
              'application/signed-exchange;v=b3;q=0.9'
}

setupResp = requests.post(url=redash_url+"/setup", data=setupPayload, headers=setupHeaders, allow_redirects=False)
print('User created')

ctJson = {'Content-Type': 'application/json;charset=UTF-8'}
datasourceQuery = {
    'options': {
        'host': os.getenv('DB_HOST', 'localhost'), 'port': int(os.getenv('DB_PORT', '5432')),
        'user': os.getenv('DB_USER', 'postgres'), 'password': os.getenv('DB_PASS'),
        'dbname': os.getenv('DB_NAME', 'public')
    },
    'type': os.getenv('DB_TYPE', 'pg'), 'name': os.getenv('DATASOURCE_NAME', 'default')
}

if os.getenv('DB_SSL_MODE') is not None:
    datasourceQuery['options']['sslmode'] = os.getenv('DB_SSL_MODE')

# TODO: Save datasource id
datasourceResp = requests.post(url=redash_url+"/api/data_sources", cookies=setupResp.cookies, json=datasourceQuery,
                               headers=ctJson)
print('Datasource created')

usersResp = requests.get(url=redash_url+"/api/users/1", cookies=setupResp.cookies)
apiKey = usersResp.json()['api_key']
print('Api key:', apiKey)

redash = RedashAPIClient(apiKey, redash_url)

dashboard_name = os.getenv('DASHBOARD_NAME')
dashboard_resp = redash.create_dashboard(dashboard_name)
dashboard_id = dashboard_resp.json()['id']
print('Created dashboard', dashboard_name)

for file_name in glob.iglob('./*.json', recursive=True):
    f = open(file_name, "r")
    widget_json = f.read()
    if len(widget_json) > 0:
        widget = json.loads(widget_json)
        widget['dashboard_id'] = dashboard_id
        widget_resp = redash.post('widgets', widget)
        print('Created widget from', file_name)

for file_name in glob.iglob('./*.sql', recursive=True):
    f = open(file_name, "r")

    query_name = f.readline()[2:].strip()
    query_description = f.readline()[2:].strip()
    visualization = json.loads(f.readline()[2:].strip())
    widget_json = f.readline()[2:].strip()
    widget = {}
    if len(widget_json) > 3:
        widget = json.loads(widget_json)
    query = f.read()

    query_resp = redash.create_query(ds_id=1, name=query_name, qry=query, desc=query_description)
    query_id = query_resp.json()['id']
    print('Created query', query_name, 'id:', query_id)

    if len(visualization) > 3:
        visualization['query_id'] = query_id
        vis_resp = redash.post('visualizations', visualization)
        vis_id = vis_resp.json()['id']
        print('Created visualisation for', query_name, 'query. Visualization id:', vis_id)

        redash.generate_query_results(ds_id=1, qry=query, qry_id=query_id)
        print('Generated query results for', query_name, 'query.')

        publish_resp = requests.post(url="{}{}{}{}".format(redash_url, "/queries", query_id, "/source"),
                                     cookies=setupResp.cookies, headers=ctJson,
                                     data={'id': query_id, 'version': query_resp.json()['version'], 'is_draft': False})

        if len(widget_json) > 3:
            widget['dashboard_id'] = dashboard_id
            widget['visualization_id'] = vis_id
            widget_resp = redash.post('widgets', widget)
            print('Created widget for', query_name, 'query')

redash.publish_dashboard(dashboard_id)
print('Published dashboard', dashboard_name)




