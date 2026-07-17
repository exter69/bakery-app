<template>
  <div class="bg-white">
    <div class="mx-auto max-w-2xl px-4 py-16 sm:px-6 sm:py-24 lg:max-w-7xl lg:px-8">
      <h2 class="sr-only">Products</h2>

      <div class="grid grid-cols-1 gap-y-4 sm:grid-cols-2 sm:gap-x-6 sm:gap-y-10 lg:grid-cols-3 lg:gap-x-8">
        <div v-for="product in products" :key="product.id" class="group relative flex flex-col overflow-hidden rounded-lg border border-gray-200 bg-white">
          <img :src="product.imageSrc" :alt="product.imageAlt" class="aspect-3/4 w-full bg-gray-200 object-cover group-hover:opacity-75 sm:aspect-auto sm:h-96" />
          <div class="flex flex-1 flex-col space-y-2 p-4">
            <h3 class="text-sm font-medium text-gray-900">
              <a :href="product.href">
                <span aria-hidden="true" class="absolute inset-0"></span>
                {{ product.name }}
              </a>
            </h3>
            <p class="text-sm text-gray-500">{{ product.description }}</p>
            <div class="flex flex-1 flex-col justify-end">
              <p class="text-sm text-gray-500 italic">{{ product.options }}</p>
              <p class="text-base font-medium text-gray-900">{{ product.price }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const products = [
  {
    id: 1,
    name: 'Basic Tee 8-Pack',
    href: '#',
    price: '$256',
    description: 'Get the full lineup of our Basic Tees. Have a fresh shirt all week, and an extra for laundry day.',
    options: '8 colors',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/category-page-02-image-card-01.jpg',
    imageAlt: 'Eight shirts arranged on table in black, olive, grey, blue, white, red, mustard, and green.',
  },
  {
    id: 2,
    name: 'Basic Tee',
    href: '#',
    price: '$32',
    description: 'Look like a visionary CEO and wear the same black t-shirt every day.',
    options: 'Black',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/category-page-02-image-card-02.jpg',
    imageAlt: 'Front of plain black t-shirt.',
  },
  {
    id: 3,
    name: 'Kinda White Basic Tee',
    href: '#',
    price: '$32',
    description: "It's probably, like, 5000 Kelvin instead of 6000 K.",
    options: 'White',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/category-page-02-image-card-03.jpg',
    imageAlt: 'Front of plain white t-shirt.',
  },
  {
    id: 4,
    name: 'Stone Basic Tee',
    href: '#',
    price: '$32',
    description: 'White tees stain easily, and black tees fade. This is going to be gray for a while.',
    options: 'Charcoal',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/category-page-02-image-card-04.jpg',
    imageAlt: 'Front of plain dark gray t-shirt.',
  },
  {
    id: 5,
    name: 'Fall Basic Tee 3-Pack',
    href: '#',
    price: '$96',
    description: 'Who need stark minimalism when you could have earth tones? Embrace the season.',
    options: 'Charcoal',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/category-page-02-image-card-05.jpg',
    imageAlt: 'Three shirts arranged on table in mustard, dark gray, and olive.',
  },
  {
    id: 6,
    name: 'Linework Artwork Tee 3-Pack',
    href: '#',
    price: '$108',
    description: 'Get all 3 colors of our popular Linework design and some variety to your monotonous life.',
    options: '3 colors',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/category-page-02-image-card-06.jpg',
    imageAlt:
      'Three shirts in gray, white, and blue arranged on table with same line drawing of hands and shapes overlapping on front of shirt.',
  },
]
</script>