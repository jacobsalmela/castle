package ui

import (
	"game/assets"
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/resources"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// RenderHeadbars draws enemy health bars using the ECS render queue.
// Each entity with HeadbarData and Transform is rendered if it should be visible.
//
// Note: This system is currently dormant and will be activated when map rendering
// migrates to the queue-based system (Phase 6), ensuring proper layer composition.
func RenderHeadbars(world *ecs.World, queue *resources.RenderQueue) {
	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Query all entities with HeadbarData and Transform components
	entityIDs := world.EntitiesWith((*components.HeadbarData)(nil), (*components.Transform)(nil))

	for _, entityID := range entityIDs {
		headbar := ecs.GetComponent[components.HeadbarData](world, entityID)
		transform := ecs.GetComponent[components.Transform](world, entityID)

		if headbar == nil || transform == nil {
			continue
		}

		// Only render if the headbar should be visible
		// Show when: ShowTimer > 0 && Health < MaxHealth && Health > 0
		if !(headbar.ShowTimer > 0 && headbar.Health < headbar.MaxHealth && headbar.Health > 0) {
			continue
		}

		// Calculate bar position relative to entity
		barW := float64(assets.HeadbarBarImage.Bounds().Dx())
		barX := (headbar.EntityWidth - barW) / 2
		barY := -7.0

		// Create transform for the bar (apply entity position, then bar offset)
		barTransform := ebiten.GeoM{}
		barTransform.Translate(transform.X, transform.Y)
		barTransform.Translate(barX, barY)

		// Draw outer border on normal map (subtract from normals)
		normalGeoM := barTransform
		queue.Push(resources.RenderCommand{
			Image:      assets.HeadbarBarImage,
			TargetType: resources.TargetNormal,
			Layer:      config.PipelineUILayer,
			GeoM:       normalGeoM,
			ColorScale: components.NormalMaskColor,
		})

		// Draw outer border on screen
		queue.Push(resources.RenderCommand{
			Image:      assets.HeadbarBarImage,
			TargetType: resources.TargetScreen,
			Layer:      config.PipelineUILayer,
			GeoM:       barTransform,
		})

		// Draw inner empty background (offset 1px inside border)
		innerTransform := barTransform
		innerTransform.Translate(1, 1)
		queue.Push(resources.RenderCommand{
			Image:      assets.HeadbarInnerBarImage,
			TargetType: resources.TargetScreen,
			Layer:      config.PipelineUILayer,
			GeoM:       innerTransform,
		})

		// Draw lag bar (yellow) if there's lag damage
		if lagWidth := math.Floor((headbar.HealthLag / headbar.MaxHealth) * float64(cfg.Hud.EnemyBarW)); lagWidth > 0 {
			lagGeoM := ebiten.GeoM{}
			lagGeoM.Scale(lagWidth, 1)
			lagGeoM.Concat(innerTransform)
			queue.Push(resources.RenderCommand{
				Image:      assets.HeadbarLagBarImage,
				TargetType: resources.TargetScreen,
				Layer:      config.PipelineUILayer,
				GeoM:       lagGeoM,
			})
		}

		// Draw current health bar (red)
		if healthWidth := math.Round((headbar.Health / headbar.MaxHealth) * float64(cfg.Hud.EnemyBarW)); healthWidth > 0 {
			healthGeoM := ebiten.GeoM{}
			healthGeoM.Scale(healthWidth, 1)
			healthGeoM.Concat(innerTransform)
			queue.Push(resources.RenderCommand{
				Image:      assets.HeadbarFillerBarImage,
				TargetType: resources.TargetScreen,
				Layer:      config.PipelineUILayer,
				GeoM:       healthGeoM,
			})
		}
	}
}

// RenderHeadbar renders a single entity's headbar given its EntityId.
// This is a convenience function for rendering individual headbars.
func RenderHeadbar(world *ecs.World, queue *resources.RenderQueue, entityID entities.EntityId) {
	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	headbar := ecs.GetComponent[components.HeadbarData](world, entityID)
	transform := ecs.GetComponent[components.Transform](world, entityID)

	if headbar == nil || transform == nil {
		return
	}

	// Check if headbar should be visible
	// Show when: ShowTimer > 0 && Health < MaxHealth && Health > 0
	if !(headbar.ShowTimer > 0 && headbar.Health < headbar.MaxHealth && headbar.Health > 0) {
		return
	}

	// Calculate bar position
	barW := float64(assets.HeadbarBarImage.Bounds().Dx())
	barX := (headbar.EntityWidth - barW) / 2
	barY := -7.0

	barTransform := ebiten.GeoM{}
	barTransform.Translate(transform.X, transform.Y)
	barTransform.Translate(barX, barY)

	// Normal map punch-out
	queue.Push(resources.RenderCommand{
		Image:      assets.HeadbarBarImage,
		TargetType: resources.TargetNormal,
		Layer:      config.PipelineUILayer,
		GeoM:       barTransform,
		ColorScale: components.NormalMaskColor,
	})

	// Outer border
	queue.Push(resources.RenderCommand{
		Image:      assets.HeadbarBarImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       barTransform,
	})

	// Inner background
	innerTransform := barTransform
	innerTransform.Translate(1, 1)
	queue.Push(resources.RenderCommand{
		Image:      assets.HeadbarInnerBarImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       innerTransform,
	})

	// Lag bar
	if lagWidth := math.Floor((headbar.HealthLag / headbar.MaxHealth) * float64(cfg.Hud.EnemyBarW)); lagWidth > 0 {
		lagGeoM := ebiten.GeoM{}
		lagGeoM.Scale(lagWidth, 1)
		lagGeoM.Concat(innerTransform)
		queue.Push(resources.RenderCommand{
			Image:      assets.HeadbarLagBarImage,
			TargetType: resources.TargetScreen,
			Layer:      config.PipelineUILayer,
			GeoM:       lagGeoM,
		})
	}

	// Health bar
	if healthWidth := math.Round((headbar.Health / headbar.MaxHealth) * float64(cfg.Hud.EnemyBarW)); healthWidth > 0 {
		healthGeoM := ebiten.GeoM{}
		healthGeoM.Scale(healthWidth, 1)
		healthGeoM.Concat(innerTransform)
		queue.Push(resources.RenderCommand{
			Image:      assets.HeadbarFillerBarImage,
			TargetType: resources.TargetScreen,
			Layer:      config.PipelineUILayer,
		})
	}
}

// RenderHeadbarAtPosition renders a headbar at the given camera-relative position.
// This is used by the UI system to render headbars without requiring HeadbarData components.
func RenderHeadbarAtPosition(world *ecs.World, queue *resources.RenderQueue, data *components.HeadbarData, x, y float64) {
	if queue == nil || data == nil {
		return
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Check if headbar should be visible
	// Show when: ShowTimer > 0 && Health < MaxHealth && Health > 0
	if !(data.ShowTimer > 0 && data.Health < data.MaxHealth && data.Health > 0) {
		return
	}

	// Calculate bar position
	barW := float64(assets.HeadbarBarImage.Bounds().Dx())
	barX := x + (data.EntityWidth-barW)/2
	barY := y - 7.0

	barTransform := ebiten.GeoM{}
	barTransform.Translate(barX, barY)

	// Normal map punch-out
	queue.Push(resources.RenderCommand{
		Image:      assets.HeadbarBarImage,
		TargetType: resources.TargetNormal,
		Layer:      config.PipelineUILayer,
		GeoM:       barTransform,
		ColorScale: components.NormalMaskColor,
	})

	// Outer border
	queue.Push(resources.RenderCommand{
		Image:      assets.HeadbarBarImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       barTransform,
	})

	// Inner background
	innerTransform := barTransform
	innerTransform.Translate(1, 1)
	queue.Push(resources.RenderCommand{
		Image:      assets.HeadbarInnerBarImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       innerTransform,
	})

	// Lag bar
	if lagWidth := math.Floor((data.HealthLag / data.MaxHealth) * float64(cfg.Hud.EnemyBarW)); lagWidth > 0 {
		lagGeoM := ebiten.GeoM{}
		lagGeoM.Scale(lagWidth, 1)
		lagGeoM.Concat(innerTransform)
		queue.Push(resources.RenderCommand{
			Image:      assets.HeadbarLagBarImage,
			TargetType: resources.TargetScreen,
			Layer:      config.PipelineUILayer,
			GeoM:       lagGeoM,
		})
	}

	// Health bar
	if healthWidth := math.Round((data.Health / data.MaxHealth) * float64(cfg.Hud.EnemyBarW)); healthWidth > 0 {
		healthGeoM := ebiten.GeoM{}
		healthGeoM.Scale(healthWidth, 1)
		healthGeoM.Concat(innerTransform)
		queue.Push(resources.RenderCommand{
			Image:      assets.HeadbarFillerBarImage,
			TargetType: resources.TargetScreen,
			Layer:      config.PipelineUILayer,
			GeoM:       healthGeoM,
		})
	}
}
