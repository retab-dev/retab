type SchemaTypeInfer<T extends string> =
    T extends "string" ? string :
    T extends "number" ? number :
    T extends "integer" ? number :
    T extends "boolean" ? boolean :
    T extends "array" ? any[] :
    T extends "object" ? Record<string, any> :
    T extends "null" ? null :
    never;

type BuildAnyOf<T> = T extends [infer U, ...infer Rest] ? SchemaInfer<U> | BuildAnyOf<Rest> : never;
type BuildAllOf<T> = T extends [infer U, ...infer Rest] ? SchemaInfer<U> & BuildAllOf<Rest> : any;

// @ts-ignore
type WalkProp<Ctx, Prop> = Ctx[Prop];
type ResolveRelativeRef<Ctx, Ref extends string> =
    Ref extends `${infer Prop}/${infer Rest}` ?
        ResolveRelativeRef<WalkProp<Ctx, Prop>, Rest> :
        WalkProp<Ctx, Ref>;
    
type ResolveRef<Ctx, Ref extends string> =
    Ref extends `#/${infer Rest}` ? ResolveRelativeRef<Ctx, Rest> : never;

type BuildTypes<T extends string[]> = T extends [infer Type extends string, ...infer Rest extends string[]] ? SchemaTypeInfer<Type> | BuildTypes<Rest> : never;

type SchemaInfer<T, Ctx = T> =
    T extends {$ref: string} ? ResolveRef<Ctx, T["$ref"]> :
    T extends {anyOf: any[]} ? BuildAnyOf<T["anyOf"]> :
    T extends {allOf: any[]} ? BuildAllOf<T["allOf"]> :
    T extends {type: string[]} ? BuildTypes<T["type"]> :
    T extends {type: string} ? SchemaTypeInfer<T["type"]> :
    never;
