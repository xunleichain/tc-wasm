(module
 (type $FUNCSIG$ii (func (param i32) (result i32)))
 (type $FUNCSIG$iii (func (param i32 i32) (result i32)))
 (import "env" "malloc" (func $malloc (param i32) (result i32)))
 (import "env" "strcpy" (func $strcpy (param i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 16) "localhost\00")
 (data (i32.const 32) "test.onething.com\00")
 (export "memory" (memory $0))
 (export "thunderchain_main" (func $thunderchain_main))
 (func $thunderchain_main (; 2 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (local $3 i32)
  (drop
   (call $strcpy
    (tee_local $2
     (call $malloc
      (i32.const 128)
     )
    )
    (i32.const 16)
   )
  )
  (drop
   (call $strcpy
    (tee_local $3
     (i32.add
      (get_local $2)
      (i32.const 64)
     )
    )
    (i32.const 32)
   )
  )
  (i32.store8 offset=64
   (get_local $2)
   (i32.const 70)
  )
  (get_local $3)
 )
)