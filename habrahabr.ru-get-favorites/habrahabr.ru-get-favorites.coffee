#!/usr/bin/env coffee

# get user favorites from habrahabr.ru and geektimes.ru

Jsdom = require "jsdom"

HABRAHABR_HOSTS = ["http://habrahabr.ru", "http://geektimes.ru"]
HABR_USER_NAME = process.argv[2] || process.env.USER

# ------------------------------------------------------------------
get_favorites_from_page = (habrahabr_host, favorites_url, result, on_complete_result) ->
    favorites_url = "#{habrahabr_host}/users/#{HABR_USER_NAME}/favorites/" if favorites_url == ""
    Jsdom.env
        url: favorites_url
        scripts: ["https://code.jquery.com/jquery-2.1.3.min.js"]
        done: (errors, window) ->
            $ = window.jQuery

            $('h1.title a.post_title').each (i, el) ->
                result.push
                    title: el.textContent
                    href: el.href

            next_page = $('div.page-nav > ul > li a[id=next_page]')
            if next_page.length > 0
                get_favorites_from_page habrahabr_host, next_page[0].href, result, on_complete_result
            else
                on_complete_result habrahabr_host, result

# ------------------------------------------------------------------
for habrahabr_host in HABRAHABR_HOSTS
    get_favorites_from_page habrahabr_host, "", [], (result_host, result) ->
        console.log "#{result_host} favorites for #{HABR_USER_NAME}:"
        result.reverse().map (row) ->
            console.log "#{row.href} #{row.title}"
        console.log ""
