import sqlite3
import datetime

def convert_array_to_string(data):
    def inner(t):
        if isinstance(t, str):
            return '"' + t + '"'
        return str(t)


    if isinstance(data, list):
        return "{" + ",".join([inner(t) for t in data])  + "}"
    return data


old_conn = sqlite3.connect('media-provider.db.old')
old_cursor = old_conn.cursor()

new_conn = sqlite3.connect('media-provider.db')
new_cursor = new_conn.cursor()

old_cursor.execute('SELECT * FROM main.users')
users = old_cursor.fetchall()
now = datetime.datetime.now(datetime.UTC)
new_cursor.executemany('INSERT INTO main.users (id, created_at, updated_at, name, password_hash, api_key, permission, original) VALUES (?, ?, ?, ?, ?, ?, ?, ?);',
                       [(user[0], now, now, user[1], user[2], user[3], user[4], user[5]) for user in users])

old_cursor.execute('SELECT * FROM main.subscriptions')
subscriptions = old_cursor.fetchall()
new_cursor.executemany('INSERT INTO main.subscriptions (id, created_at, updated_at, provider, content_id, refresh_frequency) VALUES (? , ?, ?, ?, ?, ?)',
                       [(sub[0], now, now, sub[1], sub[2], sub[3]) for sub in subscriptions])

old_cursor.execute('SELECT * FROM main.subscription_info')
subscriptions_info = old_cursor.fetchall()
new_cursor.executemany('INSERT INTO main.subscription_infos (subscription_id, created_at, updated_at,  title, description, last_check_success, base_dir, last_check) VALUES (?, ?, ?, ?, ?, ?, ?, ?)',
                       [(si[0], now, now, si[1], si[2], si[3], si[4], si[5]) for si in subscriptions_info])

old_cursor.execute('SELECT * FROM main.pages')
pages = old_cursor.fetchall()
old_cursor.execute('SELECT * FROM providers')
providers = old_cursor.fetchall()
old_cursor.execute('SELECT * FROM dirs')
dirs = old_cursor.fetchall()

new_cursor.executemany('INSERT INTO main.pages (id, created_at, updated_at, title, sort_value, providers, dirs, custom_root_dir) VALUES (?, ?, ?, ?, ?, ?, ?, ?);',
                       [(p[0], now, now, p[1], p[3],
                         convert_array_to_string([prov[1] for prov in providers if prov[0] == p[0]]),
                         convert_array_to_string([d[1] for d in dirs if d[0] == p[0]]),
                         p[2]
                         ) for p in pages])


old_cursor.execute('SELECT * FROM main.modifiers')
modifiers = old_cursor.fetchall()
new_cursor.executemany(
    '''INSERT INTO main.modifiers (id, created_at, updated_at, page_id, title, type, key) 
    VALUES (?, ?, ?, ?, ?, ?, ?)''',
    [(row[0], now, now, row[1], row[2], row[3], row[4]) for row in modifiers if row[0] > 0]
)

old_cursor.execute('SELECT * FROM main.modifier_values')
modifier_values = old_cursor.fetchall()
new_cursor.executemany(
    '''INSERT INTO main.modifier_values (created_at, updated_at, modifier_id, key, value) 
    VALUES (?, ?, ?, ?, ?)''',
    [(now, now, row[0], row[1], row[2]) for row in modifier_values]
)


new_conn.commit()

old_conn.close()
new_conn.close()

