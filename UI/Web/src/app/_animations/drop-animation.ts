import {animate, query, sequence, stagger, style, transition, trigger} from "@angular/animations";


export const dropAnimation = trigger("dropDown", [
  transition(":enter", [
    style({ height: 0, overflow: "hidden" }),
    query(".dropdown-item", [
      style({ opacity: 0, transform: "translateY(-50px)" })
    ]),
    sequence([
      animate("200ms", style({ height: "*" })),
      query(".dropdown-item", [
        stagger(-50, [
          animate("400ms ease", style({ opacity: 1, transform: "none" }))
        ])
      ])
    ])
  ]),

  transition(":leave", [
    style({ height: "*", overflow: "hidden" }),
    query(".dropdown-item", [style({ opacity: 1, transform: "none" })]),
    sequence([
      query(".dropdown-item", [
        stagger(50, [
          animate(
            "400ms ease",
            style({ opacity: 0, transform: "translateY(-50px)" })
          )
        ])
      ]),
      animate("200ms", style({ height: 0 }))
    ])
  ])
])
