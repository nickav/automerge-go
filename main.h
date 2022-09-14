#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

#include "include/automerge.h"

AMdoc *toAMdoc(AMvalue value)
{
    return value.doc;
}

AMbyteSpan toAMbyteSpan(AMvalue value)
{
    return value.bytes;
}

AMobjId *toAMobjId(AMvalue value)
{
    return (AMobjId *)value.obj_id;
}

AMdoc *AMG_Create(AMactorId *actor_id)
{
    return AMresultValue(AMcreate(actor_id)).doc;
}

void AMG_Commit(AMdoc *doc, const char *msg, const time_t *time)
{
    AMfree(AMcommit(doc, msg, time));
}

size_t AMG_Rollback(AMdoc *doc)
{
    size_t result = 0;
    if (doc != NULL)
    {
        result = AMrollback(doc);
    }
    return result;
}

AMbyteSpan AMG_Save(AMdoc *doc)
{
    AMbyteSpan result = {};
    if (doc != NULL)
    {
        // @Incomplete: copy bytes and free AMresult
        AMresult *memory_leak = AMsave(doc);
        result = AMresultValue(memory_leak).bytes;
    }
    return result;
}

AMdoc *AMG_Load(const char *src, size_t count)
{
    // @MemoryLeak: AMresult needs to be freed
    AMresult *result = AMload((uint8_t *)src, count);
    return AMresultValue(result).doc;
}

void AMG_Merge(AMdoc *dest, AMdoc *src)
{
    AMfree(AMmerge(dest, src));
}

AMactorId *AMG_GetActorID(AMdoc *doc)
{
    AMresult *result = AMgetActorId(doc);
    return (AMactorId *)AMresultValue(result).actor_id;
}

const char *AMG_GetActorIDString(AMdoc *doc)
{
    AMactorId *actor_id = AMG_GetActorID(doc);
    return AMactorIdStr(actor_id);
}

void AMG_MapPutObject(AMdoc *doc, const char *key)
{
}