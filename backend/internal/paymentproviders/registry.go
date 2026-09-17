package paymentproviders

import "fmt"

type Factory func(Config) PaymentProvider

type Descriptor struct {
	Name         ProviderName `json:"name"`
	DisplayName  string       `json:"display_name"`
	Implemented  bool         `json:"implemented"`
	Capabilities []string     `json:"capabilities"`
}

type Registry struct {
	factories   map[ProviderName]Factory
	descriptors map[ProviderName]Descriptor
}

func NewRegistry() *Registry {
	registry := &Registry{factories: make(map[ProviderName]Factory), descriptors: make(map[ProviderName]Descriptor)}
	registry.Register(ProviderCardcom, Descriptor{
		Name: ProviderCardcom, DisplayName: "Cardcom", Implemented: true,
		Capabilities: []string{"create", "verify", "cancel", "refund", "webhook"},
	}, func(config Config) PaymentProvider { return NewCardcomAdapter(config) })
	registry.RegisterDescriptor(ProviderPayMe, "PayMe", false)
	registry.RegisterDescriptor(ProviderGrow, "Grow", false)
	registry.RegisterDescriptor(ProviderPalPay, "PalPay", false)
	registry.RegisterDescriptor(ProviderStripe, "Stripe", false)
	registry.RegisterDescriptor(ProviderPayPal, "PayPal", false)
	return registry
}

func (r *Registry) Register(name ProviderName, descriptor Descriptor, factory Factory) {
	descriptor.Name = name
	descriptor.Implemented = factory != nil
	r.descriptors[name] = descriptor
	if factory != nil {
		r.factories[name] = factory
	}
}

func (r *Registry) RegisterDescriptor(name ProviderName, displayName string, implemented bool) {
	r.descriptors[name] = Descriptor{Name: name, DisplayName: displayName, Implemented: implemented}
}

func (r *Registry) Build(name ProviderName, config Config) (PaymentProvider, error) {
	factory, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("provider %q is not implemented or configured", name)
	}
	return factory(config), nil
}

func (r *Registry) Descriptors() []Descriptor {
	result := make([]Descriptor, 0, len(r.descriptors))
	for _, descriptor := range r.descriptors {
		result = append(result, descriptor)
	}
	return result
}
