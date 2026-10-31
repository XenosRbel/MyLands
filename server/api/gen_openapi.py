import sys, re, json
# (method path, tag, summary, auth: none|user|admin, idempotent command?, body fields)
S = [
("GET /healthz","ops","Liveness","none",False,None),
("GET /readyz","ops","Readiness (DB ping)","none",False,None),
("GET /metrics","ops","Prometheus metrics (text)","none",False,None),
("GET /api/v1/time","ops","Server time","none",False,None),
("GET /packs/{pack}/assets/{file...}","ops","Static pack asset","none",False,None),
("POST /api/v1/auth/register","auth","Register login/password; returns session","none",False,{"login":"string","password":"string","displayName":"string?"}),
("POST /api/v1/auth/login","auth","Login; returns session","none",False,{"login":"string","password":"string"}),
("POST /api/v1/auth/guest","auth","Create guest account; returns session","none",False,{}),
("POST /api/v1/auth/platform","auth","Sign in with a platform identity (only 'dev' implemented; others 501)","none",False,{"platform":"string","externalId":"string","signature":"string?"}),
("POST /api/v1/auth/logout","auth","Revoke current token","user",False,{}),
("POST /api/v1/auth/upgrade","auth","Attach login/password to guest account","user",False,{"login":"string","password":"string"}),
("GET /api/v1/me","auth","Current account","user",False,None),
("PUT /api/v1/me/locale","auth","Set preferred locale","user",False,{"locale":"string"}),
("DELETE /api/v1/me","auth","Delete (anonymise) account","user",False,None),
("GET /api/v1/worlds","worlds","List worlds with join state","user",False,None),
("GET /api/v1/worlds/{wid}/pack","worlds","Pinned Game Pack JSON (ETag = pack hash)","user",False,None),
("POST /api/v1/worlds/{wid}/join","worlds","Create player + capital","user",False,{"name":"string","race":"string"}),
("GET /api/v1/worlds/{wid}/overview","player","Player overview (cities, counters, server time)","user",False,None),
("GET /api/v1/worlds/{wid}/quests","quests","Quest list with status/progress","user",False,None),
("POST /api/v1/worlds/{wid}/quests/{qid}/claim","quests","Claim quest reward (once)","user",True,{"cityId":"integer"}),
("GET /api/v1/worlds/{wid}/missions","missions","Own missions + incoming to own cities","user",False,None),
("POST /api/v1/worlds/{wid}/missions/{mid}/recall","missions","Recall outbound or stationed mission","user",True,{}),
("GET /api/v1/worlds/{wid}/heroes","heroes","Heroes, inventory, limits","user",False,None),
("POST /api/v1/worlds/{wid}/heroes/{hid}/skills","heroes","Learn skill level","user",True,{"skill":"string"}),
("POST /api/v1/worlds/{wid}/heroes/{hid}/equip","heroes","Equip item","user",True,{"itemId":"integer"}),
("POST /api/v1/worlds/{wid}/heroes/{hid}/unequip","heroes","Unequip slot","user",True,{"slot":"string"}),
("GET /api/v1/worlds/{wid}/territories","territories","Owned territories with pending yield","user",False,None),
("POST /api/v1/worlds/{wid}/territories/abandon","territories","Abandon territory","user",True,{"x":"integer","y":"integer"}),
("GET /api/v1/worlds/{wid}/reports","reports","Reports (cursor: before, limit)","user",False,None),
("GET /api/v1/worlds/{wid}/reports/{rid}","reports","Report detail (marks read)","user",False,None),
("GET /api/v1/worlds/{wid}/notifications","notifications","Notifications (after, limit) + unread count","user",False,None),
("POST /api/v1/worlds/{wid}/notifications/read","notifications","Mark notifications read up to id","user",True,{"upTo":"integer"}),
("GET /api/v1/worlds/{wid}/messages","messages","Messages (box=inbox|sent|alliance, before, limit)","user",False,None),
("POST /api/v1/worlds/{wid}/messages","messages","Send message or alliance broadcast","user",True,{"to":"string","alliance":"boolean","subject":"string","body":"string"}),
("POST /api/v1/worlds/{wid}/messages/{id}/read","messages","Mark message read","user",True,{}),
("POST /api/v1/worlds/{wid}/messages/{id}/report","messages","Report message for moderation","user",True,{}),
("POST /api/v1/worlds/{wid}/blocks","messages","Block/unblock a player","user",True,{"playerId":"integer","block":"boolean"}),
("GET /api/v1/worlds/{wid}/map","map","Map viewport (x,y,w,h ≤ 64)","user",False,None),
("GET /api/v1/worlds/{wid}/rankings","map","Player rankings (limit, offset)","user",False,None),
("GET /api/v1/worlds/{wid}/market","market","Open orders (sell, buy, mine=1)","user",False,None),
("POST /api/v1/worlds/{wid}/market/{oid}/cancel","market","Cancel own order (refund escrow)","user",True,{}),
("GET /api/v1/worlds/{wid}/alliances","alliances","Alliance list","user",False,None),
("GET /api/v1/worlds/{wid}/alliances/{aid}","alliances","Alliance detail (members, relations)","user",False,None),
("POST /api/v1/worlds/{wid}/alliances","alliances","Found alliance","user",True,{"name":"string","tag":"string","description":"string"}),
("GET /api/v1/worlds/{wid}/alliance-invites","alliances","Pending invites (received / sent by my alliance)","user",False,None),
("POST /api/v1/worlds/{wid}/alliance/invites","alliances","Invite player by name (leader/officer)","user",True,{"player":"string"}),
("POST /api/v1/worlds/{wid}/alliance-invites/{iid}/respond","alliances","Accept/decline invite","user",True,{"accept":"boolean"}),
("POST /api/v1/worlds/{wid}/alliance/leave","alliances","Leave alliance (last member disbands)","user",True,{}),
("POST /api/v1/worlds/{wid}/alliance/members/{pid}/role","alliances","Set role leader|officer|member or kick","user",True,{"role":"string"}),
("POST /api/v1/worlds/{wid}/alliance/relations","alliances","Set diplomacy ally|nap|war|neutral","user",True,{"allianceId":"integer","status":"string"}),
("GET /api/v1/cities/{cid}","city","City view: resources, rates, buildings/techs/units with next steps, queues","user",False,None),
("POST /api/v1/cities/{cid}/build","city","Queue next building level","user",True,{"building":"string"}),
("POST /api/v1/cities/{cid}/research","city","Queue next technology level","user",True,{"tech":"string"}),
("POST /api/v1/cities/{cid}/train","city","Train units","user",True,{"unit":"string","quantity":"integer"}),
("POST /api/v1/cities/{cid}/jobs/cancel","city","Cancel last job of a queue (partial refund)","user",True,{"jobId":"integer"}),
("POST /api/v1/cities/{cid}/rename","city","Rename city","user",True,{"name":"string"}),
("POST /api/v1/cities/{cid}/missions","missions","Send mission (attack|scout|transport|reinforce|settle|occupy)","user",True,{"kind":"string","x":"integer","y":"integer","units":"object","heroId":"integer?","cargo":"object?","capture":"boolean?","retreatPct":"integer?"}),
("POST /api/v1/cities/{cid}/heroes","heroes","Hire hero","user",True,{"class":"string","name":"string"}),
("POST /api/v1/cities/{cid}/market/orders","market","Place order (escrow)","user",True,{"sell":"string","sellAmount":"integer","buy":"string","buyAmount":"integer"}),
("POST /api/v1/cities/{cid}/market/accept","market","Accept order","user",True,{"orderId":"integer"}),
("POST /api/v1/cities/{cid}/exchange","market","NPC merchant exchange","user",True,{"from":"string","to":"string","amount":"integer"}),
("POST /api/v1/cities/{cid}/collect","territories","Collect territory yield into city","user",True,{"x":"integer","y":"integer"}),
("GET /api/v1/admin/players/{pid}","admin","Inspect player (settled), pending events, ledger (audited)","admin",False,None),
("GET /api/v1/admin/events/failed","admin","Parked events","admin",False,None),
("POST /api/v1/admin/events/{eid}/retry","admin","Re-queue parked event (audited)","admin",False,{}),
("POST /api/v1/admin/cities/{cid}/grant","admin","Grant/remove resources with reason (audited)","admin",False,{"amounts":"object","reason":"string"}),
("POST /api/v1/admin/accounts/{aid}/ban","admin","Ban account until time (audited)","admin",False,{"until":"string","reason":"string"}),
("GET /api/v1/admin/audit","admin","Admin audit log","admin",False,None),
("GET /api/v1/admin/messages/reported","admin","Reported messages","admin",False,None),
("POST /api/v1/admin/worlds/{wid}/migrate-pack","admin","Migrate world to new pack version (dry-run supported, audited)","admin",False,{"packDir":"string","dryRun":"boolean"}),
]
out=[]
w=out.append
w("openapi: 3.1.0")
w("info:")
w("  title: Realmforge Game API")
w("  version: 0.1.0")
w("  description: |")
w("    Server-authoritative REST API. Envelope: success `{\"data\": ...}`, failure `{\"error\": {\"code\", \"message\", \"details?\"}}`.")
w("    State-changing game commands (marked x-idempotent) REQUIRE an `Idempotency-Key` header (8-80 chars [A-Za-z0-9_-]);")
w("    duplicates replay the stored response byte-for-byte with header `Idempotent-Replay: true`; reusing a key for a different body returns 422.")
w("    Generated by server/api/gen_openapi.py; route coverage is enforced by internal/api/openapi_test.go.")
w("servers:\n  - url: http://localhost:8080")
w("components:")
w("  securitySchemes:\n    bearer:\n      type: http\n      scheme: bearer")
w("  parameters:\n    IdempotencyKey:\n      name: Idempotency-Key\n      in: header\n      required: true\n      schema: {type: string, pattern: '^[A-Za-z0-9_-]{8,80}$'}")
w("  schemas:")
w("    Error:\n      type: object\n      properties:\n        error:\n          type: object\n          required: [code, message]\n          properties:\n            code: {type: string}\n            message: {type: string}\n            details: {}")
w("    Data:\n      type: object\n      properties:\n        data: {}")
w("  responses:")
for code,desc in [("BadRequest","Invalid request / JSON / missing idempotency key"),("Unauthorized","Missing or invalid token"),("Forbidden","Not allowed"),("NotFound","Not found (also returned for other players' resources)"),("Conflict","State conflict (e.g. already claimed, order closed)"),("Unprocessable","Game rule violation (e.g. insufficient_resources) or idempotency key reuse"),("TooMany","Rate limited")]:
    w(f"    {code}:\n      description: {desc}\n      content:\n        application/json:\n          schema: {{$ref: '#/components/schemas/Error'}}")
w("paths:")
bypath={}
for (mp,tag,summ,auth,idem,body) in S:
    m,p=mp.split(" ",1)
    bypath.setdefault(p,[]).append((m.lower(),tag,summ,auth,idem,body))
for p in bypath:
    w(f"  {p}:")
    for (m,tag,summ,auth,idem,body) in bypath[p]:
        w(f"    {m}:")
        w(f"      tags: [{tag}]")
        w(f"      summary: {json.dumps(summ, ensure_ascii=False)}")
        params=re.findall(r"\{([a-z]+)(?:\.\.\.)?\}",p)
        plist=[f"        - {{name: {x}, in: path, required: true, schema: {{type: {'string' if x in ('qid','pack','file') else 'integer'}}}}}" for x in params]
        if idem: plist.append("        - $ref: '#/components/parameters/IdempotencyKey'")
        if plist:
            w("      parameters:"); out.extend(plist)
        if auth=="none": w("      security: []")
        else: w("      security: [{bearer: []}]")
        if auth=="admin": w("      x-requires-role: admin")
        if idem: w("      x-idempotent: true")
        if body is not None:
            w("      requestBody:\n        required: true\n        content:\n          application/json:\n            schema:\n              type: object\n              additionalProperties: false")
            req=[k for k,v in body.items() if not v.endswith("?")]
            if req: w(f"              required: [{', '.join(req)}]")
            if body:
                w("              properties:")
                for k,v in body.items():
                    w(f"                {k}: {{type: {v.rstrip('?')}}}")
        w("      responses:")
        w("        '200':\n          description: OK\n          content:\n            application/json:\n              schema: {$ref: '#/components/schemas/Data'}")
        for c,r in [("400","BadRequest"),("401","Unauthorized"),("403","Forbidden"),("404","NotFound"),("409","Conflict"),("422","Unprocessable"),("429","TooMany")]:
            if auth=="none" and c in ("401","403"): continue
            w(f"        '{c}': {{$ref: '#/components/responses/{r}'}}")
open(sys.argv[1],"w").write("\n".join(out)+"\n")
print(len(S),"operations")
