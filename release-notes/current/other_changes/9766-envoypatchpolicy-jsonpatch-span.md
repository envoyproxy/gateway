Added a dedicated `Translator.processJSONPatches` tracing span so the time spent
applying EnvoyPatchPolicy JSON patches can be isolated from the rest of the xDS
translation.
