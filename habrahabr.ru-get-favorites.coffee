#!/usr/bin/env coffee

# get user favorites from habrahabr.ru

Jsdom = require "jsdom"

habrahabr_host = "http://habrahabr.ru"
habr_user_name = process.argv[2] || process.env.USER

console.log "habrahabr.ru favorites for #{habr_user_name}"

# ------------------------------------------------------------------
get_favorites_from_page = (address, result, on_complete_result) ->
    Jsdom.env
        url: address
        scripts: ["http://code.jquery.com/jquery-2.0.3.min.js"]
        done: (errors, window) ->
            $ = window.jQuery

            $('h1.title a.post_title').each (i, el) ->
                result.push
                    title: el.textContent
                    href: el.href

            next_page = $('div.page-nav > ul > li a[id=next_page]')
            if next_page.length > 0
                get_favorites_from_page next_page[0].href, result, on_complete_result
            else
                on_complete_result result

# ------------------------------------------------------------------
get_favorites_from_page "#{habrahabr_host}/users/#{habr_user_name}/favorites/", [], (result) ->
    result.reverse().map (row) ->
        console.log "#{row.href} #{row.title}"
