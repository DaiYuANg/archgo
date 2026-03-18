package dix

func Invoke0(fn func()) InvokeFunc { return func(c *Container) error { dixInvoke0(c, fn); return nil } }
func Invoke1[T any](fn func(T)) InvokeFunc {
	return func(c *Container) error { return dixInvoke1(c, fn) }
}
func Invoke2[T1, T2 any](fn func(T1, T2)) InvokeFunc {
	return func(c *Container) error { return dixInvoke2(c, fn) }
}
func Invoke3[T1, T2, T3 any](fn func(T1, T2, T3)) InvokeFunc {
	return func(c *Container) error { return dixInvoke3(c, fn) }
}
func Invoke4[T1, T2, T3, T4 any](fn func(T1, T2, T3, T4)) InvokeFunc {
	return func(c *Container) error { return dixInvoke4(c, fn) }
}
func Invoke5[T1, T2, T3, T4, T5 any](fn func(T1, T2, T3, T4, T5)) InvokeFunc {
	return func(c *Container) error { return dixInvoke5(c, fn) }
}
func Invoke6[T1, T2, T3, T4, T5, T6 any](fn func(T1, T2, T3, T4, T5, T6)) InvokeFunc {
	return func(c *Container) error { return dixInvoke6(c, fn) }
}

func dixInvoke0(c *Container, fn func()) { fn() }
func dixInvoke1[T any](c *Container, fn func(T)) error {
	t, err := ResolveAs[T](c)
	if err != nil {
		return err
	}
	fn(t)
	return nil
}
func dixInvoke2[T1, T2 any](c *Container, fn func(T1, T2)) error {
	t1, err := ResolveAs[T1](c)
	if err != nil {
		return err
	}
	t2, err := ResolveAs[T2](c)
	if err != nil {
		return err
	}
	fn(t1, t2)
	return nil
}
func dixInvoke3[T1, T2, T3 any](c *Container, fn func(T1, T2, T3)) error {
	t1, err := ResolveAs[T1](c)
	if err != nil {
		return err
	}
	t2, err := ResolveAs[T2](c)
	if err != nil {
		return err
	}
	t3, err := ResolveAs[T3](c)
	if err != nil {
		return err
	}
	fn(t1, t2, t3)
	return nil
}
func dixInvoke4[T1, T2, T3, T4 any](c *Container, fn func(T1, T2, T3, T4)) error {
	t1, err := ResolveAs[T1](c)
	if err != nil {
		return err
	}
	t2, err := ResolveAs[T2](c)
	if err != nil {
		return err
	}
	t3, err := ResolveAs[T3](c)
	if err != nil {
		return err
	}
	t4, err := ResolveAs[T4](c)
	if err != nil {
		return err
	}
	fn(t1, t2, t3, t4)
	return nil
}
func dixInvoke5[T1, T2, T3, T4, T5 any](c *Container, fn func(T1, T2, T3, T4, T5)) error {
	t1, err := ResolveAs[T1](c)
	if err != nil {
		return err
	}
	t2, err := ResolveAs[T2](c)
	if err != nil {
		return err
	}
	t3, err := ResolveAs[T3](c)
	if err != nil {
		return err
	}
	t4, err := ResolveAs[T4](c)
	if err != nil {
		return err
	}
	t5, err := ResolveAs[T5](c)
	if err != nil {
		return err
	}
	fn(t1, t2, t3, t4, t5)
	return nil
}
func dixInvoke6[T1, T2, T3, T4, T5, T6 any](c *Container, fn func(T1, T2, T3, T4, T5, T6)) error {
	t1, err := ResolveAs[T1](c)
	if err != nil {
		return err
	}
	t2, err := ResolveAs[T2](c)
	if err != nil {
		return err
	}
	t3, err := ResolveAs[T3](c)
	if err != nil {
		return err
	}
	t4, err := ResolveAs[T4](c)
	if err != nil {
		return err
	}
	t5, err := ResolveAs[T5](c)
	if err != nil {
		return err
	}
	t6, err := ResolveAs[T6](c)
	if err != nil {
		return err
	}
	fn(t1, t2, t3, t4, t5, t6)
	return nil
}
