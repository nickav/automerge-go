#include <stdio.h>
#include <stdlib.h>
#include <assert.h>

#include "include/automerge.h"

AMdoc *AMG_Create(AMactorId *actor_id)
{
    return AMresultValue(AMcreate(actor_id)).doc;
}

AMdoc *AMG_Clone(AMdoc *doc)
{
    return AMresultValue(AMclone(doc)).doc;
}

bool AMG_Equals(AMdoc *a, AMdoc *b)
{
    bool result = false;
    if (a != NULL && b != NULL)
    {
        result = AMequal(a, b);
    }
    return result;
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

const char *AMG_ActorIDToString(AMactorId *actor_id)
{
    return AMactorIdStr(actor_id);
}

const char *AMG_GetActorIDString(AMdoc *doc)
{
    AMactorId *actor_id = AMG_GetActorID(doc);
    return AMactorIdStr(actor_id);
}

AMvalue AMG_MapGet(AMdoc *doc, AMobjId *obj_id, const char *key)
{
    // @MemoryLeak
    return AMresultValue(AMmapGet(doc, obj_id, key, NULL));
}

// @Incomplete: return if the key was deleted or not
// @Robustness: key _must_ exist in the object according to the automerge docs
void AMG_MapDelete(AMdoc *doc, AMobjId *obj_id, const char *key)
{
    if (doc != NULL)
    {
        AMfree(AMmapDelete(doc, obj_id, key));
    }
}

AMvalue AMG_ListGet(AMdoc *doc, AMobjId *obj_id, size_t index)
{
    // @MemoryLeak
    return AMresultValue(AMlistGet(doc, obj_id, index, NULL));
}

// @Incomplete: return if the key was deleted or not
// @Robustness: key _must_ exist in the object according to the automerge docs
void AMG_ListDelete(AMdoc *doc, AMobjId *obj_id, size_t index)
{
    if (doc != NULL)
    {
        AMfree(AMlistDelete(doc, obj_id, index));
    }
}

size_t AMG_GetSize(AMdoc *doc, AMobjId *obj_id)
{
    return AMobjSize(doc, obj_id, NULL);
}

AMvalueVariant AMG_GetType(AMvalue value)
{
    return value.tag;
}

int AMG_ToBool(AMvalue value)
{
    return (int)value.boolean;
}

int64_t AMG_ToInt(AMvalue value)
{
    return (int64_t)value.int_;
}

double AMG_ToF64(AMvalue value)
{
    return value.f64;
}

const char *AMG_ToString(AMvalue value)
{
    return value.str;
}

AMactorId *AMG_ToActorID(AMvalue value)
{
    return (AMactorId *)value.actor_id;
}

uint64_t AMG_ToUint(AMvalue value)
{
    return value.uint;
}

AMobjId *AMG_ToObjectID(AMvalue value)
{
    return (AMobjId *)value.obj_id;
}

AMdoc *AMG_ToDocument(AMvalue value)
{
    return value.doc;
}

AMstrs AMG_ToStrs(AMvalue value)
{
    return value.strs;
}
