#!/usr/bin/env coffee

# get user favorites from habrahabr.ru and geektimes.ru
# usage:
#     install node.js & coffee script
#     ./habr.com-get-favorites.coffee habr_user_name > habr.com-favorites.txt

Jsdom = require "jsdom"

HABRAHABR_HOSTS = ["https://habr.com", "https://geektimes.com"]
HABR_USER_NAME = process.argv[2] || process.env.USER

# ------------------------------------------------------------------
get_favorites_from_page = (habrahabr_host, favorites_url, result, on_complete_result) ->
    favorites_url = "#{habrahabr_host}/users/#{HABR_USER_NAME}/favorites/" if favorites_url == ""
    Jsdom.env
        url: favorites_url
        scripts: ["https://code.jquery.com/jquery-3.3.1.slim.min.js"]
        done: (errors, window) ->
            $ = window.jQuery

            $('h2.post__title a.post__title_link').each (i, el) ->
                result.push
                    title: el.textContent
                    href: el.href

            next_page = $('div.page__footer > ul > li a[id=next_page]')
            if next_page.length > 0
                get_favorites_from_page habrahabr_host, next_page[0].href, result, on_complete_result
            else
                on_complete_result habrahabr_host, result

# ------------------------------------------------------------------
for habrahabr_host in HABRAHABR_HOSTS
    get_favorites_from_page habrahabr_host, "", [], (result_host, result) ->
        console.log "#{result_host} favorites for #{HABR_USER_NAME}:"
        console.log "-----------------------------------------------"
        result.reverse().map (row) ->
            console.log "#{row.href} #{row.title}"
        console.log ""
