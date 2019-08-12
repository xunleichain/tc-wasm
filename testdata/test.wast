(module
 (type $FUNCSIG$viii (func (param i32 i32 i32)))
 (import "env" "ONT_JsonUnmashalInput" (func $ONT_JsonUnmashalInput (param i32 i32 i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "invoke" (func $invoke))
 (func $invoke (; 1 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (i32.store offset=4
   (i32.const 0)
   (tee_local $2
    (i32.sub
     (i32.load offset=4
      (i32.const 0)
     )
     (i32.const 16)
    )
   )
  )
  (call $ONT_JsonUnmashalInput
   (i32.add
    (get_local $2)
    (i32.const 8)
   )
   (i32.const 8)
   (get_local $1)
  )
  (i32.store offset=4
   (i32.const 0)
   (i32.add
    (get_local $2)
    (i32.const 16)
   )
  )
  (i32.const 0)
 )
)
